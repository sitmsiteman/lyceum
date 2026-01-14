package tlgcore

import (
	"bytes"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
	"unicode"
)

var greekBase = map[rune]rune{
	'a': 'α', 'b': 'β', 'g': 'γ', 'd': 'δ', 'e': 'ε', 'z': 'ζ', 'h': 'η', 'q': 'θ',
	'i': 'ι', 'k': 'κ', 'l': 'λ', 'm': 'μ', 'n': 'ν', 'c': 'ξ', 'o': 'ο', 'p': 'π',
	'r': 'ρ', 's': 'σ', 'j': 'ς', 't': 'τ', 'u': 'υ', 'f': 'φ', 'x': 'χ', 'y': 'ψ', 'w': 'ω',
}

var diacritics = map[rune]string{
	')': "\u0313", '(': "\u0314", '/': "\u0301", '\\': "\u0300",
	'=': "\u0342", '+': "\u0308", '|': "\u0345",
}
var alphaBase = map[rune]string{
	'α': "a", 'β': "b", 'γ': "g", 'δ': "d", 'ε': "e", 'ζ': "z", 'η': "h", 'θ': "q",
	'ι': "i", 'κ': "k", 'λ': "l", 'μ': "m", 'ν': "n", 'ξ': "c", 'ο': "o", 'π': "p",
	'ρ': "r", 'σ': "s", 'ς': "s", 'τ': "t", 'υ': "u", 'φ': "f", 'χ': "x", 'ψ': "y", 'ω': "w",
	'ά': "a/", 'έ': "e/", 'ή': "h/", 'ί': "i/", 'ό': "o/", 'ύ': "u/", 'ώ': "w/",
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
		out.WriteString("Title: ")
	case "{70": // TLG Editorial Text
		isLatin = true
	default:
		isLatin = false
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

	// Cleanup artifacts
	//s = regexp.MustCompile(`@\{.*?\}`).ReplaceAllString(s, "")
	//s = strings.ReplaceAll(s, "-", " ")
	//s = strings.ReplaceAll(s, "1", "")

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
				continue
			}

			if c, ok := greekBase[unicode.ToLower(r)]; ok {
				if upper {
					out.WriteRune(unicode.ToUpper(c))
					upper = false
				} else {
					out.WriteRune(c)
				}
				continue
			} else if d, ok := diacritics[r]; ok {
				out.WriteString(d)
				continue
			}
		}

		out.WriteRune(r)
	}
	res := out.String()
	// Hacky but works
	res = regexp.MustCompile(`σ(\s|[[:punct:]]|$)`).ReplaceAllString(res, "ς$1")
	res = norm.NFC.String(res)
	return res
}

func ToBetaCode(s string) string {
	var out strings.Builder
	for _, r := range s {
		lower := unicode.ToLower(r)
		if val, ok := alphaBase[lower]; ok {
			out.WriteString(val)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
