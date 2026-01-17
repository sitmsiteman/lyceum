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

	'A': 'Α', 'B': 'Β', 'G': 'Γ', 'D': 'Δ', 'E': 'Ε', 'Z': 'Ζ', 'H': 'Η', 'Q': 'Θ',
	'I': 'Ι', 'K': 'Κ', 'L': 'Λ', 'M': 'Μ', 'N': 'Ν', 'C': 'Ξ', 'O': 'Ο', 'P': 'Π',
	'R': 'Ρ', 'S': 'Σ', 'J': 'Σ', 'T': 'Τ', 'U': 'Υ', 'F': 'Φ', 'X': 'Χ', 'Y': 'Ψ', 'W': 'Ω',
}

var diacritics = map[rune]string{
	')': "\u0313", '(': "\u0314", '/': "\u0301", '\\': "\u0300",
	'=': "\u0342", '+': "\u0308", '|': "\u0345",
}

var alphaBase = map[rune]string{
	'α': "a", 'β': "b", 'γ': "g", 'δ': "d", 'ε': "e", 'ζ': "z", 'η': "h", 'θ': "q",
	'ι': "i", 'κ': "k", 'λ': "l", 'μ': "m", 'ν': "n", 'ξ': "c", 'ο': "o", 'π': "p",
	'ρ': "r", 'σ': "s", 'ς': "s", 'τ': "t", 'υ': "u", 'φ': "f", 'χ': "x", 'ψ': "y", 'ω': "w",

	'ά': "a/", 'ὰ': "a\\", 'ᾶ': "a=",
	'ἀ': "a)", 'ἄ': "a)/", 'ἂ': "a)\\", 'ἆ': "a)=",
	'ἁ': "a(", 'ἅ': "a(/", 'ἃ': "a(\\", 'ἇ': "a(=",
	'ᾳ': "a|",
	'ᾴ': "a/|", 'ᾲ': "a\\|", 'ᾷ': "a=|",
	'ᾀ': "a)|", 'ᾄ': "a)/|", 'ᾂ': "a)\\|", 'ᾆ': "a)=|",
	'ᾁ': "a(|", 'ᾅ': "a(/|", 'ᾃ': "a(\\|", 'ᾇ': "a(=|",
	'ᾱ': "a%", 'ᾰ': "a&",

	'έ': "e/", 'ὲ': "e\\",
	'ἐ': "e)", 'ἔ': "e)/", 'ἒ': "e)\\",
	'ἑ': "e(", 'ἕ': "e(/", 'ἓ': "e(\\",

	'ή': "h/", 'ὴ': "h\\", 'ῆ': "h=",
	'ἠ': "h)", 'ἤ': "h)/", 'ἢ': "h)\\", 'ἦ': "h)=",
	'ἡ': "h(", 'ἥ': "h(/", 'ἣ': "h(\\", 'ἧ': "h(=",
	'ῃ': "h|",
	'ῄ': "h/|", 'ῂ': "h\\|", 'ῇ': "h=|",
	'ᾐ': "h)|", 'ᾔ': "h)/|", 'ᾒ': "h)\\|", 'ᾖ': "h)=|",
	'ᾑ': "h(|", 'ᾕ': "h(/|", 'ᾓ': "h(\\|", 'ᾗ': "h(=|",

	'ί': "i/", 'ὶ': "i\\", 'ῖ': "i=",
	'ἰ': "i)", 'ἴ': "i)/", 'ἲ': "i)\\", 'ἶ': "i)=",
	'ἱ': "i(", 'ἵ': "i(/", 'ἳ': "i(\\", 'ἷ': "i(=",
	'ϊ': "i+", 'ΐ': "i+/", 'ῒ': "i+\\",
	'ῑ': "i%", 'ῐ': "i&",
	'ό': "o/", 'ὸ': "o\\",
	'ὀ': "o)", 'ὄ': "o)/", 'ὂ': "o)\\",
	'ὁ': "o(", 'ὅ': "o(/", 'ὃ': "o(\\",

	'ύ': "u/", 'ὺ': "u\\", 'ῦ': "u=",
	'ὐ': "u)", 'ὔ': "u)/", 'ὒ': "u)\\", 'ὖ': "u)=",
	'ὑ': "u(", 'ὕ': "u(/", 'ὓ': "u(\\", 'ὗ': "u(=",
	'ϋ': "u+", 'ΰ': "u+/", 'ῢ': "u+\\", 'ῧ': "u+=",
	'ῡ': "u%", 'ῠ': "u&",

	'ώ': "w/", 'ὼ': "w\\", 'ῶ': "w=",
	'ὠ': "w)", 'ὤ': "w)/", 'ὢ': "w)\\", 'ὦ': "w)=",
	'ὡ': "w(", 'ὥ': "w(/", 'ὣ': "w(\\", 'ὧ': "w(=",
	'ῳ': "w|",
	'ῴ': "w/|", 'ῲ': "w\\|", 'ῷ': "w=|",
	'ᾠ': "w)|", 'ᾤ': "w)/|", 'ᾢ': "w)\\|", 'ᾦ': "w)=|",
	'ᾡ': "w(|", 'ᾥ': "w(/|", 'ᾣ': "w(\\|", 'ᾧ': "w(=|",

	'ῤ': "r)", 'ῥ': "r(",

	'Α': "*a", 'Β': "*b", 'Γ': "*g", 'Δ': "*d", 'Ε': "*e", 'Ζ': "*z", 'Η': "*h", 'Θ': "*q",
	'Ι': "*i", 'Κ': "*k", 'Λ': "*l", 'Μ': "*m", 'Ν': "*n", 'Ξ': "*c", 'Ο': "*o", 'Π': "*p",
	'Ρ': "*r", 'Σ': "*s", 'Τ': "*t", 'Υ': "*u", 'Φ': "*f", 'Χ': "*x", 'Ψ': "*y", 'Ω': "*w",

	'Ά': "*a/", 'Ὰ': "*a\\",
	'Ἀ': "*a)", 'Ἄ': "*a)/", 'Ἂ': "*a)\\", 'Ἆ': "*a)=",
	'Ἁ': "*a(", 'Ἅ': "*a(/", 'Ἃ': "*a(\\", 'Ἇ': "*a(=",
	'ᾼ': "*a|",
	'ᾈ': "*a)|", 'ᾌ': "*a)/|", 'ᾊ': "*a)\\|", 'ᾎ': "*a)=|",
	'ᾉ': "*a(|", 'ᾍ': "*a(/|", 'ᾋ': "*a(\\|", 'ᾏ': "*a(=|",
	'Ᾱ': "*a%", 'Ᾰ': "*a&",

	'Έ': "*e/", 'Ὲ': "*e\\",
	'Ἐ': "*e)", 'Ἔ': "*e)/", 'Ἒ': "*e)\\",
	'Ἑ': "*e(", 'Ἕ': "*e(/", 'Ἓ': "*e(\\",

	'Ή': "*h/", 'Ὴ': "*h\\",
	'Ἠ': "*h)", 'Ἤ': "*h)/", 'Ἢ': "*h)\\", 'Ἦ': "*h)=",
	'Ἡ': "*h(", 'Ἥ': "*h(/", 'Ἣ': "*h(\\", 'Ἧ': "*h(=",
	'ῌ': "*h|",
	'ᾘ': "*h)|", 'ᾜ': "*h)/|", 'ᾚ': "*h)\\|", 'ᾞ': "*h)=|",
	'ᾙ': "*h(|", 'ᾝ': "*h(/|", 'ᾛ': "*h(\\|", 'ᾟ': "*h(=|",

	'Ί': "*i/", 'Ὶ': "*i\\",
	'Ἰ': "*i)", 'Ἴ': "*i)/", 'Ἲ': "*i)\\", 'Ἶ': "*i)=",
	'Ἱ': "*i(", 'Ἵ': "*i(/", 'Ἳ': "*i(\\", 'Ἷ': "*i(=",
	'Ϊ': "*i+",
	'Ῑ': "*i%", 'Ῐ': "*i&",

	'Ό': "*o/", 'Ὸ': "*o\\",
	'Ὀ': "*o)", 'Ὄ': "*o)/", 'Ὂ': "*o)\\",
	'Ὁ': "*o(", 'Ὅ': "*o(/", 'Ὃ': "*o(\\",

	'Ύ': "*u/", 'Ὺ': "*u\\",
	'὘': "*u)",
	'Ὑ': "*u(", 'Ὕ': "*u(/", 'Ὓ': "*u(\\", 'Ὗ': "*u(=",
	'Ϋ': "*u+",
	'Ῡ': "*u%", 'Ῠ': "*u&",

	'Ώ': "*w/", 'Ὼ': "*w\\",
	'Ὠ': "*w)", 'Ὤ': "*w)/", 'Ὢ': "*w)\\", 'Ὦ': "*w)=",
	'Ὡ': "*w(", 'Ὥ': "*w(/", 'Ὣ': "*w(\\", 'Ὧ': "*w(=",
	'ῼ': "*w|",
	'ᾨ': "*w)|", 'ᾬ': "*w)/|", 'ᾪ': "*w)\\|", 'ᾮ': "*w)=|",
	'ᾩ': "*w(|", 'ᾭ': "*w(/|", 'ᾫ': "*w(\\|", 'ᾯ': "*w(=|",

	'Ῥ': "*r(",

	'·': ":", ';': "?", '’': "'",
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
		if val, ok := alphaBase[r]; ok {
			out.WriteString(val)
			continue
		}

		lower := unicode.ToLower(r)
		if val, ok := alphaBase[lower]; ok {
			out.WriteString(val)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
