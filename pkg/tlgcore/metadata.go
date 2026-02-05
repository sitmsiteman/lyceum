package tlgcore

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type CitationDef struct {
	LevelChar string // "v", "w", "x", "y", "z"
	Label     string // e.g. "Book", "Line"
}

type WorkMetadata struct {
	ID        string
	Title     string
	Citations []CitationDef
}

func cleanString(s string) string {
	if strings.Contains(s, "*") {
		return ToGreek(s)
	}
	return ToLatin(s)
}

func decodeSimpleASCII(b []byte) string {
	var sb strings.Builder
	for i := 0; i < len(b); i++ {
		if b[i] == 0xFF {
			break
		}
		if b[i] >= 0x80 {
			val := b[i] & 0x7F
			if (val >= '0' && val <= '9') || (val >= 'A' && val <= 'Z') || (val >= 'a' && val <= 'z') {
				sb.WriteByte(val)
			}
		}
	}
	res := sb.String()
	if i, err := strconv.Atoi(res); err == nil {
		return strconv.Itoa(i)
	}
	return res
}

func GetAuthorName(path, tlgID string) string {
	var prefixID string
	data, err := os.ReadFile(path)
	if err != nil {
		return "Unknown"
	}

	if len(tlgID) >= 3 {
		prefixID = strings.ToUpper(tlgID[:3])
	} else {
		return "Unknown"
	}

	cleanID := fmt.Sprintf("%s%04s", prefixID, strings.TrimPrefix(strings.ToUpper(tlgID), prefixID))
	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?&1(.*?)&`, cleanID))
	matches := re.FindSubmatch(data)
	if len(matches) > 1 {
		return strings.TrimSpace(strings.Split(string(matches[1]), "&")[0])
	}
	return tlgID
}

func ReadIDT(path string) (map[string]*WorkMetadata, error) {
	m := make(map[string]*WorkMetadata)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pos := 0
	var currentWork *WorkMetadata

	// [수정] Work ID의 이전 상태를 추적 (델타 업데이트 지원)
	lastWorkIDInt := 0
	lastWorkIDStr := ""

	consumeID := func() []byte {
		start := pos
		for pos < len(data) && data[pos] >= 0x80 {
			pos++
		}
		return data[start:pos]
	}

	for pos < len(data) {
		typ := data[pos]
		pos++

		switch typ {
		case 0:
			continue
		case 1: // New Author
			if pos+4 > len(data) {
				break
			}
			pos += 4
			consumeID()
			lastWorkIDInt = 0
			lastWorkIDStr = ""
			currentWork = nil

		case 2: // New Work
			if pos+4 > len(data) { 
				break 
			}
			pos += 4
			idBytes := consumeID()

					if len(idBytes) == 0 {
				lastWorkIDInt++
				lastWorkIDStr = ""
			} else {
				lastWorkIDInt, lastWorkIDStr = DecodeWorkID(lastWorkIDInt, lastWorkIDStr, idBytes)
			}

			idStr := lastWorkIDStr
			if lastWorkIDInt != 0 {
				idStr = strconv.Itoa(lastWorkIDInt) + lastWorkIDStr
			}
			
			if idStr == "" || idStr == "0" {
				if lastWorkIDInt == 0 { lastWorkIDInt = 1 }
				idStr = strconv.Itoa(lastWorkIDInt)
			}

			currentWork = &WorkMetadata{ID: idStr}
			m[idStr] = currentWork

		case 3:
			pos += 2
		case 8, 9, 10, 12, 13, 11:
			if typ == 11 {
				pos += 2
			}
			consumeID()
		case 16: // Title
			subtype := data[pos]
			pos += 2
			length := int(data[pos-1])
			if pos+length > len(data) {
				break
			}
			str := string(data[pos : pos+length])
			pos += length
			if subtype == 1 {
				if currentWork != nil && currentWork.Title != "" {
					lastWorkIDInt++
					lastWorkIDStr = ""
					idStr := strconv.Itoa(lastWorkIDInt)
					currentWork = &WorkMetadata{ID: idStr}
					m[idStr] = currentWork
				}

				if currentWork != nil {
					currentWork.Title = cleanString(str)
				}
			}
	
		case 17: // Citations
			subtype := data[pos]
			pos += 2
			length := int(data[pos-1])
			if pos+length > len(data) {
				break
			}
			str := string(data[pos : pos+length])
			pos += length
			if currentWork != nil {
				levelChar := ""
				switch subtype {
				case 4:
					levelChar = "v"
				case 3:
					levelChar = "w"
				case 2:
					levelChar = "x"
				case 1:
					levelChar = "y"
				case 0:
					levelChar = "z"
				}
				if levelChar != "" {
					currentWork.Citations = append(currentWork.Citations, CitationDef{levelChar, cleanString(str)})
				}
			}
		default:
			continue
		}
	}
	return m, nil
}

func DecodeWorkID(prevInt int, prevStr string, b []byte) (int, string) {
	if len(b) == 0 {
		return prevInt, prevStr
	}
	if len(b) >= 2 && b[0] == 0xEF && b[1] == 0x81 {
		res := decodeSimpleASCII(b[2:])
		if i, err := strconv.Atoi(res); err == nil {
			return i, ""
		}
		return 0, res
	}

	pos := 0
	currInt, currStr := prevInt, prevStr

	readBin := func(n int) int {
		v := 0
		for i := 0; i < n && pos < len(b); i++ {
			v = (v << 7) | int(b[pos]&0x7F)
			pos++
		}
		return v
	}
	readStr := func() string {
		var sb strings.Builder
		for pos < len(b) {
			if b[pos] == 0xFF {
				pos++
				break
			}
			sb.WriteByte(b[pos] & 0x7F)
			pos++
		}
		return sb.String()
	}

	for pos < len(b) {
		val := b[pos]
		pos++
		right := val & 0x0F

		var dInt int = -999
		var dStr string
		var hasStr bool

		switch right {
		case 0x0:
			dInt = -1
		case 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7:
			dInt = int(right)
		case 0x8:
			dInt = readBin(1)
		case 0x9:
			dInt = readBin(1)
			dStr = string(rune(readBin(1)))
			hasStr = true
		case 0xA:
			dInt = readBin(1)
			dStr = readStr()
			hasStr = true
		case 0xB:
			dInt = readBin(2)
		case 0xC:
			dInt = readBin(2)
			dStr = string(rune(readBin(1)))
			hasStr = true
		case 0xD:
			dInt = readBin(2)
			dStr = readStr()
			hasStr = true
		case 0xE:
			dStr = string(rune(readBin(1)))
			hasStr = true
		case 0xF:
			dStr = readStr()
			hasStr = true
		}

			if dInt == -1 {
				currInt++
				currStr = ""
			} else if dInt != -999 {
				currInt = dInt
				if !hasStr {
					currStr = ""
				}
			}
			if hasStr {
				currStr = dStr
			}
	}
	return currInt, currStr
}
