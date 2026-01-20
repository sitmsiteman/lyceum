package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Simple scanner to debug ID parsing
func main() {
	fPath := flag.String("f", "", "File path")
	flag.Parse()

	if *fPath == "" {
		fmt.Println("Usage: debug_dump -f filename")
		return
	}

	data, err := os.ReadFile(*fPath)
	if err != nil {
		// Try uppercase
		if strings.HasSuffix(*fPath, ".txt") {
			alt := strings.TrimSuffix(*fPath, ".txt") + ".TXT"
			data, err = os.ReadFile(alt)
		}
	}
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fmt.Printf("--- Debugging %s (First 8192 bytes) ---\n", *fPath)

	limit := 8192
	if len(data) < limit {
		limit = len(data)
	}

	pos := 0

	// Helper to read bits
	readBin := func(n int) int {
		v := 0
		for i := 0; i < n; i++ {
			if pos >= limit {
				break
			}
			v = (v << 7) | int(data[pos]&0x7F)
			pos++
		}
		return v
	}

	readStr := func() string {
		var sb strings.Builder
		for pos < limit {
			b := data[pos]
			if b == 0xFF {
				pos++
				break
			}
			sb.WriteByte(b & 0x7F)
			pos++
		}
		return sb.String()
	}

	readChar := func() string {
		if pos < limit {
			b := data[pos] & 0x7F
			pos++
			return string(rune(b))
		}
		return ""
	}

	for pos < limit {
		b := data[pos]

		// If Text
		if b&0x80 == 0 {
			// Print a snippet of text and stop dumping IDs for this run
			start := pos
			for pos < limit && data[pos]&0x80 == 0 {
				pos++
			}
			txt := string(data[start:pos])
			fmt.Printf("TEXT: [%s]\n", strings.ReplaceAll(txt, "\n", "\\n"))
			continue
		}

		// If ID
		pos++
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
		case 0xE:
			if pos >= limit {
				break
			}
			next := data[pos] & 0x7F
			pos++
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
		}

		valStr := ""
		switch right {
		case 0x0:
			valStr = "INC"
		case 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7:
			valStr = fmt.Sprintf("INT=%d", right)
		case 0x8:
			valStr = fmt.Sprintf("BIN1=%d", readBin(1))
		case 0x9:
			valStr = fmt.Sprintf("BIN1=%d CHAR=%s", readBin(1), readChar())
		case 0xA:
			valStr = fmt.Sprintf("BIN1=%d STR=%s", readBin(1), readStr())
		case 0xB:
			valStr = fmt.Sprintf("BIN2=%d", readBin(2))
		case 0xC:
			valStr = fmt.Sprintf("BIN2=%d CHAR=%s", readBin(2), readChar())
		case 0xD:
			valStr = fmt.Sprintf("BIN2=%d STR=%s", readBin(2), readStr())
		case 0xE:
			valStr = fmt.Sprintf("CHAR=%s", readChar())
		case 0xF:
			if left == 0xF {
				valStr = "SPECIAL"
			} else {
				valStr = fmt.Sprintf("STR=%s", readStr())
			}
		}

		fmt.Printf("OFFSET %d: Op=%02X Level=%s Val=%s\n", pos, b, level, valStr)
	}
}
