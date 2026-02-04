package tlgcore

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

func getPriorDia(r rune) int {
	switch r {
	case '\u0313', '\u0314':
		return 1 // Breathing
	case '\u0308':
		return 2 // Diaeresis
	case '\u0300', '\u0301', '\u0342':
		return 3 // Accent
	case '\u0345':
		return 4 // Iota Subscript
	default:
		return 99 // Not a diacritic
	}
}

func parseCommand(runes []rune, start int) (string, int) {
	cmd := string(runes[start])
	curr := start + 1
	for curr < len(runes) && unicode.IsDigit(runes[curr]) {
		cmd += string(runes[curr])
		curr++
	}
	return cmd, curr - 1
}

type bcmHandler func(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool)

var bcmHandlers = map[rune]bcmHandler{
	'$': handleGreek,
	'&': handleLatin,
	'@': handlePageFormatting,
	'{': handleMarkupText,
	'}': handleEndMarkupText, // Do I need this?
	'<': handleTextFormatting,
	'"': handleQuotation,
	'[': handleOpenBracket,
	']': handleCloseBracket,
	'%': handleAddPunct,
	'#': handleAddChar,
}

// $
func handleGreek(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo

	switch command {
	case "$1":
	case "$2":
	case "$3":
	case "$10":
	default:
	}

	isLatin = false
	return nextIdx, isLatin, inQuot
}

// &
func handleLatin(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo

	switch command {
	case "&":
	default:
	}

	isLatin = true
	return nextIdx, isLatin, inQuot
}

// @
func handlePageFormatting(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "@":
		out.WriteString("  ")
	case "@6":
	case "@70":
		out.WriteString(" << ")
	case "@71":
		out.WriteString(" >> ")
	default:
	}

	return nextIdx, isLatin, inQuot
}

// {
func handleMarkupText(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "{":
		out.WriteString(" ")
	case "{70": // TLG Editorial Text
		isLatin = true
	default:
	}

	return nextIdx, isLatin, inQuot
}

// }
func handleEndMarkupText(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	default:
	}

	return nextIdx, isLatin, inQuot
}

// <
func handleTextFormatting(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "<":
	default:
	}

	return nextIdx, isLatin, inQuot
}

// "
func handleQuotation(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "\"1":
		out.WriteString("\"")
	case "\"2":
		out.WriteString("\"")
	case "\"3":
		out.WriteString("'")
	case "\"4":
		out.WriteString("'")
	case "\"5":
		out.WriteString("'")
	case "\"6":
		if !inQuo {
			out.WriteString("«")
			inQuot = true
		} else {
			out.WriteString("»")
			inQuot = false
		}
	case "\"7":
		if !inQuo {
			out.WriteString("‹")
			inQuot = true
		} else {
			out.WriteString("›")
			inQuot = false
		}
	case "\"8":
		out.WriteString("\"")
		inQuot = false
	default:
		inQuot = false
	}

	return nextIdx, isLatin, inQuot
}

// [
func handleOpenBracket(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "[":
		out.WriteString("[")
	case "[1":
		out.WriteString("(")
	case "[2":
		out.WriteString("<")
	case "[3":
		out.WriteString("{")
	case "[4":
		out.WriteString("⟦")
	case "[5":
		out.WriteString("⌊")
	case "[6":
		out.WriteString("⌈")
	case "[7":
		out.WriteString("⌈")
	case "[8":
		out.WriteString("⌊")
	case "[9":
		out.WriteString("˙")
	default:
		out.WriteString("[")
	}

	return nextIdx, isLatin, inQuot
}

// ]
func handleCloseBracket(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "]":
		out.WriteString("]")
	case "]1":
		out.WriteString(")")
	case "]2":
		out.WriteString(">")
	case "]3":
		out.WriteString("}")
	case "]4":
		out.WriteString("⟧")
	case "]5":
		out.WriteString("⌋")
	case "]6":
		out.WriteString("⌉")
	case "]7":
		out.WriteString("⌋")
	case "]8":
		out.WriteString("⌉")
	case "]9":
		out.WriteString("˙")
	default:
		out.WriteString("]")
	}

	return nextIdx, isLatin, inQuot
}

// %
func handleAddPunct(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "%":
		out.WriteString("†")
	case "%1":
		out.WriteString("?")
	case "%2":
		out.WriteString("*")
	case "%3":
		out.WriteString("/")
	case "%4":
		out.WriteString("!")
	case "%5":
		out.WriteString("|")
	case "%6":
		out.WriteString("=")
	case "%7":
		out.WriteString("+")
	case "%8":
		out.WriteString("%")
	case "%9":
		out.WriteString("&")
	case "%10":
		out.WriteString(":")
	case "%11":
		out.WriteString("•")
	case "%12":
		out.WriteString("*")
	case "%13":
		out.WriteString("‡")
	case "%14":
		out.WriteString("§")
	case "%18":
		out.WriteString("'")
	case "%19":
		out.WriteString("-")
	case "%41":
		out.WriteString("-")
	case "%43":
		out.WriteString("×")
	case "%103":
		out.WriteString("\\")
	case "%107":
		out.WriteString("~")
	default:
	}

	return nextIdx, isLatin, inQuot
}

// #
func handleAddChar(runes []rune, start int, out *bytes.Buffer, isLat bool, inQuo bool) (newIdx int, isLatin bool, inQuot bool) {
	command, nextIdx := parseCommand(runes, start)

	inQuot = inQuo
	isLatin = isLat

	switch command {
	case "#12":
		out.WriteString("—")
	case "#13":
		out.WriteString("※")
	case "#15":
		out.WriteString(">")
	case "#17":
		out.WriteString("/")
	case "#18":
		out.WriteString("<")
	default:
	}

	return nextIdx, isLatin, inQuot
}

func ToGreek(s string) string {
	var out bytes.Buffer
	upper := false
	isLatin := false
	inQuot := false

	var pDiacritics string
	wasBase := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '`' { // skip backtick
			continue
		}

		if handler, exists := bcmHandlers[r]; exists {
			nextIdx, latinState, quotState := handler(runes, i, &out, isLatin, inQuot)
			i = nextIdx
			isLatin = latinState
			inQuot = quotState
			continue
		}

		if !isLatin {
			if r == '*' {
				upper = true
				// for uppercase, Diacritics should be prebuffered.
				wasBase = false
				continue
			}

			if c, ok := GreekBase[unicode.ToLower(r)]; ok {
				if upper {
					out.WriteRune(unicode.ToUpper(c))
					upper = false
				} else {
					out.WriteRune(c)
				}

				if pDiacritics != "" {
					out.WriteString(pDiacritics)
					pDiacritics = ""
				}
				wasBase = true
				continue
			} else if d, ok := Diacritics[r]; ok {
				if wasBase {
					out.WriteString(d)
				} else {
					pDiacritics += d
				}

				continue
			}
		}

		if pDiacritics != "" {
			out.WriteString(pDiacritics)
			pDiacritics = ""
		}
		wasBase = false

		out.WriteRune(r)
	}

	if pDiacritics != "" {
		out.WriteString(pDiacritics)
		pDiacritics = ""
	}

	res := out.String()
	res = regexp.MustCompile(`σ(\s|[[:punct:],·]|$)`).ReplaceAllString(res, "ς$1")
	res = NormalizeGreek(res)
	return res
}

func NormalizeGreek(s string) string {
	var out strings.Builder
	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]

		if getPriorDia(r) < 99 {
			var dias []rune
			j := i
			for j < n && getPriorDia(runes[j]) < 99 {
				dias = append(dias, runes[j])
				j++
			}

			if j < n {
				nextR := runes[j]
				if unicode.IsLetter(nextR) {
					composed := Compose(nextR, dias)
					out.WriteRune(composed)
					i = j
					continue
				}
			}

			out.WriteString(string(dias))
			i = j - 1
			continue
		}

		if i+1 < n && getPriorDia(runes[i+1]) < 99 {
			base := r
			var dias []rune
			j := i + 1
			for j < n && getPriorDia(runes[j]) < 99 {
				dias = append(dias, runes[j])
				j++
			}

			composed := Compose(base, dias)
			out.WriteRune(composed)
			i = j - 1
			continue
		}

		out.WriteRune(r)
	}

	return out.String()
}

func Compose(base rune, diacritics []rune) rune {
	if len(diacritics) == 0 {
		return base
	}
	sortRunes(diacritics)

	var sb strings.Builder
	sb.WriteRune(base)
	for _, d := range diacritics {
		sb.WriteRune(d)
	}

	if val, ok := UnicodeComposition[sb.String()]; ok {
		return val
	}
	return base
}

func sortRunes(r []rune) {
	for i := 1; i < len(r); i++ {
		key := r[i]
		j := i - 1
		for j >= 0 && getPriorDia(r[j]) > getPriorDia(key) {
			r[j+1] = r[j]
			j--
		}
		r[j+1] = key
	}
}

func ToLatin(s string) string {
	var out bytes.Buffer

	isLatin := true
	inQuot := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// TLG/PHI metadata uses bytes < 32 or > 127 for level updates.
		if r < 32 || r > 126 {
			continue
		}

		if r == '`' {
			continue
		}

		if handler, exists := bcmHandlers[r]; exists {
			nextIdx, _, quotState := handler(runes, i, &out, isLatin, inQuot)
			i = nextIdx
			inQuot = quotState
			isLatin = true
			continue
		}

		out.WriteRune(r)
	}

	return out.String()
}

func ToBetaCode(s string) string {
	var out strings.Builder
	for _, r := range s {
		if val, ok := AlphaBase[r]; ok {
			if r == 'ς' {
				out.WriteString("s")
				continue
			}
			out.WriteString(val)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func BetaToLower(s string) string {
	runes := []rune(s)
	var out strings.Builder

	for i := 0; i < len(runes); i++ {
		if runes[i] == '*' {
			i++
			if i >= len(runes) {
				break
			}
			var diacritics []rune
			for i < len(runes) {
				r := runes[i]
				if _, ok := Diacritics[r]; ok {
					diacritics = append(diacritics, r)
					i++
				} else {
					break
				}
			}

			if i < len(runes) {
				letter := runes[i]
				out.WriteRune(letter)
				for _, d := range diacritics {
					out.WriteRune(d)
				}
			}
			continue
		}
		out.WriteRune(runes[i])
	}
	return out.String()
}
