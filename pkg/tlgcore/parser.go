package tlgcore

import (
	"fmt"
	"io"
	"os"
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

func (p *Parser) Run(targetWorkID string, listMode bool, idtTitles map[string]string) {
	seenWorks := make(map[string]bool)
	targetInt, _ := strconv.Atoi(targetWorkID)

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

			currentID := workState.Binary
			if currentID == 0 && workState.ASCII != "" {
				if val, err := strconv.Atoi(workState.ASCII); err == nil {
					currentID = val
				}
			}
			workIDStr := strconv.Itoa(currentID)
			if currentID == 0 {
				continue
			}

			if listMode {
				if !seenWorks[workIDStr] {
					seenWorks[workIDStr] = true
					title := idtTitles[workIDStr]
					if title == "" {
						title = "(Unknown Title)"
					}
					fmt.Printf("ID %-4s | %s\n", workIDStr, title)
				}
			} else {
				if currentID == targetInt {
					output := p.ProcessText(text)
					if strings.TrimSpace(output) != "" {
						cit := p.formatCitation()
						fmt.Printf("%-10s %s\n", cit, output)
					}
				}
			}
		}
	}
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
		}
	case 0xF: // Special
		if right == 0xE {
			return true
		} // End Block
		if right == 0x0 {
			return true
		} // End File
		return false
	}

	if level == "" {
		return false
	}
	st := p.Levels[level]
	st.Active = true

	// Decode Value
	switch right {
	case 0x0:
		st.Binary++
	case 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7:
		st.Binary = int(right)
		st.ASCII = ""
	case 0x8:
		st.Binary = p.readBin(1)
		st.ASCII = ""
	case 0x9:
		st.Binary = p.readBin(1)
		st.ASCII = string(p.readChar())
	case 0xA:
		st.Binary = p.readBin(1)
		st.ASCII = p.readStr()
	case 0xB:
		st.Binary = p.readBin(2)
		st.ASCII = ""
	case 0xC:
		st.Binary = p.readBin(2)
		st.ASCII = string(p.readChar())
	case 0xD:
		st.Binary = p.readBin(2)
		st.ASCII = p.readStr()
	case 0xE:
		st.ASCII = string(p.readChar())
	case 0xF:
		st.Binary = 0
		st.ASCII = p.readStr() // ASCII only
	}

	p.resetLevels(level)
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

	// Define the order of levels to check
	// We include 'n' because it often holds the rank/offset
	order := []string{"v", "w", "n", "x", "y", "z"}
	// order := []string{"v", "w", "z"}

	for _, l := range order {
		st := p.Levels[l]
		if !st.Active {
			continue
		}

		s := st.ASCII
		if st.Binary > 0 {
			// --- Arithmetic hack ---
			// Only perform '1 + b = c' if:
			// 1. We have a single ASCII letter (a-e)
			// 2. The binary offset is small (preventing mangling)
			if len(st.ASCII) == 1 && st.ASCII[0] >= 'a' && st.ASCII[0] <= 'e' && st.Binary < 10 {
				s = string(st.ASCII[0] + byte(st.Binary))
			} else {
				// Otherwise, keep them separate (e.g., "402" + "a" = "402a")
				s = strconv.Itoa(st.Binary) + st.ASCII
			}
		}

		if s != "" {
			// Standard guard to skip the database index "1" at the start
			if len(pts) == 0 && (l == "v" || l == "w") && s == "1" {
				continue
			}
			pts = append(pts, s)
		}
	}

	if len(pts) == 0 {
		return p.Levels["z"].ASCII
	}
	return strings.Join(pts, ".")
}
