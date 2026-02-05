package tlgcore

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const BlockSize = 8192

var levelRank = map[string]int{
	"a": 0, "b": 1, "c": 2, "d": 3,
	"n": 4, "v": 5, "w": 6, "x": 7, "y": 8, "z": 9,
}

type IDState struct {
	Binary int
	ASCII  string
	Active bool
}

type Parser struct {
	File        *os.File
	Levels      map[string]*IDState
	Buffer      []byte
	Pos         int
	IsLatinFile bool

	IDTData     map[string]*WorkMetadata
	CurrentMeta *WorkMetadata

	SortedLevels []string
}

func NewParser(f *os.File) *Parser {
	p := &Parser{
		File:   f,
		Levels: make(map[string]*IDState),
		Buffer: make([]byte, BlockSize),
	}
	for k := range levelRank {
		p.Levels[k] = &IDState{}
	}
	return p
}

func (p *Parser) ProcessText(s string) string {
	if p.IsLatinFile {
		return ToLatin(s)
	}
	return ToGreek(s)
}

func (p *Parser) ResetInternalState() {
	p.File.Seek(0, 0)
	p.Pos = 0
	for k := range levelRank {
		p.Levels[k] = &IDState{}
	}
}

func (p *Parser) analyzeCitationLevels() {
	if p.CurrentMeta == nil {
		p.SortedLevels = []string{}
		return
	}

	var levels []string
	seen := make(map[string]bool)

	for _, cit := range p.CurrentMeta.Citations {
		if !seen[cit.LevelChar] {
			levels = append(levels, cit.LevelChar)
			seen[cit.LevelChar] = true
		}
	}

	sort.Slice(levels, func(i, j int) bool {
		return levelRank[levels[i]] < levelRank[levels[j]]
	})

	p.SortedLevels = levels
}

func (p *Parser) ExtractList(idtData map[string]*WorkMetadata) ([]string, error) {
	p.ResetInternalState()

	seenWorks := make(map[string]bool)
	var results []string

	for {
		n, err := p.File.Read(p.Buffer)
		if n == 0 || err == io.EOF {
			break
		}
		p.Pos = 0

		for p.Pos < n {
			b := p.Buffer[p.Pos]
			if b&0x80 != 0 {
				if p.parseIDByte() {
					break
				}
				continue
			}

			_ = p.readText(n)

			workState := p.Levels["b"]
			if !workState.Active {
				continue
			}

			currentID := p.getCurrentWorkID()
			if currentID == "0" {
				continue
			}

			if !seenWorks[currentID] {
				seenWorks[currentID] = true
				title := "(Unknown Title)"
				if meta, ok := idtData[currentID]; ok {
					title = meta.Title
				}
				line := fmt.Sprintf("ID:%-4s | %s", currentID, title)
				results = append(results, line)
			}
		}
	}
	return results, nil
}

func (p *Parser) ExtractWork(targetWorkID string) (string, error) {
	p.ResetInternalState()

	if p.IDTData != nil {
		p.CurrentMeta = p.IDTData[targetWorkID]
		p.analyzeCitationLevels()
	}

	var sb strings.Builder
	targetInt, _ := strconv.Atoi(targetWorkID)
	found := false

	for {
		n, err := p.File.Read(p.Buffer)
		if n == 0 || err == io.EOF {
			break
		}
		p.Pos = 0

		for p.Pos < n {
			b := p.Buffer[p.Pos]
			if b&0x80 != 0 {
				if p.parseIDByte() {
					break
				}
				continue
			}

			text := p.readText(n)
			if len(text) == 0 {
				continue
			}

			workState := p.Levels["b"]
			if !workState.Active {
				continue
			}

			currentID := p.getCurrentWorkID()
			currentInt := 0
			if val, err := strconv.Atoi(currentID); err == nil {
				currentInt = val
			}

			if currentID == targetWorkID || currentInt == targetInt {
				found = true
				output := p.ProcessText(text)
				if strings.TrimSpace(output) != "" {
					cit := p.formatCitation()
					sb.WriteString(fmt.Sprintf("%-10s %s\n", cit, output))
				}
			} else if found {
				return sb.String(), nil
			}
		}
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("work ID %s not found", targetWorkID)
	}

	return sb.String(), nil
}

func (p *Parser) getCurrentWorkID() string {
	st := p.Levels["b"]
	if st.Binary > 0 {
		return strconv.Itoa(st.Binary)
	}
	if st.ASCII != "" {
		// Try to parse ASCII as int for normalization (e.g. "001" -> "1")
		if val, err := strconv.Atoi(st.ASCII); err == nil {
			return strconv.Itoa(val)
		}
		return st.ASCII
	}
	return "0"
}

func (p *Parser) parseIDByte() bool {
	if p.Pos >= len(p.Buffer) {
		return true
	}
	b := p.Buffer[p.Pos]

	p.Pos++

	left := (b >> 4) & 0x0F
	right := b & 0x0F
	level := ""

	switch left {
	case 0x8:
		level = "z"
	case 0x9:
		level = "y"
	case 0xA:
		level = "x"
	case 0xB:
		level = "w"
	case 0xC:
		level = "v"
	case 0xD:
		level = "n"
	case 0xE: // Escape
		if p.Pos >= len(p.Buffer) {
			return true
		}
		next := p.Buffer[p.Pos] & 0x7F
		p.Pos++
		switch next {
		case 0:
			level = "a"
		case 1:
			level = "b"
		case 2:
			level = "c"
		case 4:
			level = "d"
		default:
		}
	case 0xF: // Special
		if right == 0xE {
			return true
		}
		if right == 0x0 {
			return true
		}
		return false
	}

	const UnsetBinary = -999
	binaryVal := UnsetBinary
	asciiVal := ""
	hasASCII := false

	switch right {
	case 0x0:
		binaryVal = -1 // Increment
	case 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7:
		binaryVal = int(right)
	case 0x8:
		binaryVal = p.readBin(1)
	case 0x9:
		binaryVal = p.readBin(1)
		asciiVal = string(p.readChar())
		hasASCII = true
	case 0xA:
		binaryVal = p.readBin(1)
		asciiVal = p.readStr()
		hasASCII = true
	case 0xB:
		binaryVal = p.readBin(2)
	case 0xC:
		binaryVal = p.readBin(2)
		asciiVal = string(p.readChar())
		hasASCII = true
	case 0xD:
		binaryVal = p.readBin(2)
		asciiVal = p.readStr()
		hasASCII = true
	case 0xE:
		asciiVal = string(p.readChar())
		hasASCII = true
	case 0xF:
		asciiVal = p.readStr()
		hasASCII = true
	}

	if level != "" {
		st := p.Levels[level]

		oldActive := st.Active
		oldBinary := st.Binary
		oldASCII := st.ASCII

		st.Active = true

		if binaryVal == -1 {
			st.Binary++
			st.ASCII = ""
		} else if binaryVal != UnsetBinary {
			st.Binary = binaryVal
			if !hasASCII {
				st.ASCII = ""
			}
		}

		if hasASCII {
			st.ASCII = asciiVal
		}

		isTwoRank := false

		if len(p.SortedLevels) == 2 {
			if p.SortedLevels[0] == level {
				isTwoRank = true
			}
		}

		if isTwoRank && oldASCII == "a" {
			if st.Binary == oldBinary+1 && st.ASCII == "" {
				st.Binary = oldBinary // 숫자를 이전 숫자로 되돌림 (25 -> 24)
				st.ASCII = "b"        // 접미사를 b로 설정
			}
		}

		if !oldActive || st.Binary != oldBinary || st.ASCII != oldASCII {
			p.resetLevels(level)
		}
	}

	return false
}

func (p *Parser) resetLevels(lvl string) {
	rank := levelRank[lvl]
	resetToNull := (lvl == "a" || lvl == "b" || lvl == "n")
	for l, r := range levelRank {
		if r > rank {
			if resetToNull {
				p.Levels[l].Binary = 0
				p.Levels[l].ASCII = ""
				p.Levels[l].Active = false
			} else {
				p.Levels[l].Binary = 1
				p.Levels[l].ASCII = ""
				p.Levels[l].Active = true
			}
		}
	}
}

// normalizeID removes leading zeros from string IDs (e.g. "001" -> "1")
func NormalizeID(id string) string {
	i, err := strconv.Atoi(id)
	if err == nil {
		return strconv.Itoa(i)
	}
	return id
}

func (p *Parser) readText(lim int) string {
	s := p.Pos
	for p.Pos < lim {
		if p.Buffer[p.Pos]&0x80 != 0 {
			break
		}
		p.Pos++
	}
	return strings.ReplaceAll(string(p.Buffer[s:p.Pos]), "\x00", "")
}

func (p *Parser) readBin(n int) int {
	v := 0
	for i := 0; i < n; i++ {
		if p.Pos >= len(p.Buffer) {
			break
		}
		v = (v << 7) | int(p.Buffer[p.Pos]&0x7F)
		p.Pos++
	}
	return v
}
func (p *Parser) readChar() rune {
	if p.Pos < len(p.Buffer) {
		b := p.Buffer[p.Pos] & 0x7F
		p.Pos++
		return rune(b)
	}
	return ' '
}
func (p *Parser) readStr() string {
	var sb strings.Builder
	for p.Pos < len(p.Buffer) {
		b := p.Buffer[p.Pos]
		if b == 0xFF {
			p.Pos++
			break
		}
		sb.WriteByte(b & 0x7F)
		p.Pos++
	}
	return sb.String()
}

func (p *Parser) formatCitation() string {
	var pts []string
	var levelsToCheck []string

	if p.CurrentMeta != nil && len(p.CurrentMeta.Citations) > 0 {
		for _, def := range p.CurrentMeta.Citations {
			levelsToCheck = append(levelsToCheck, def.LevelChar)
		}
	} else {
		levelsToCheck = []string{"w", "x", "y", "z"}
	}

	isStephanus := (len(p.SortedLevels) == 3)
	sectionLevel := ""
	if isStephanus {
		sectionLevel = p.SortedLevels[1]
	}

	seenLevel := make(map[string]bool)

	for _, l := range levelsToCheck {

		if seenLevel[l] {
			continue
		}

		st := p.Levels[l]
		if st == nil || !st.Active {
			continue
		}

		s := st.ASCII
		if st.Binary > 0 {
			if isStephanus && l == sectionLevel && st.Binary >= 1 && st.Binary <= 5 {
				if s == "" {
					s = string('a' + byte(st.Binary-1))
				}
			} else {
				s = strconv.Itoa(st.Binary) + s
			}
		}

		if s != "" {
			pts = append(pts, s)
			seenLevel[l] = true
		}
	}

	if len(pts) == 0 && p.Levels["z"].Active {
		return p.Levels["z"].ASCII
	}
	return strings.Join(pts, ".")
}
