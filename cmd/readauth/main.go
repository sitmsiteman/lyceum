package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

type Record struct {
	ID      string
	Author  string
	Epithet string
}

func main() {
	fPath := flag.String("f", "authtab.dir", "filename")
	flag.Parse()

	data, err := ioutil.ReadFile(*fPath)
	if err != nil {
		log.Fatal(err)
	}

	var records []Record
	i := 0
	for i < len(data) {
		if isNewRecordStart(data[i:]) {
			rec, nextPos := decodeEntry(data, i)
			records = append(records, rec)
			i = nextPos
			continue
		}
		i++
	}

	for _, r := range records {
		fmt.Printf("%-8s | %s %s\n", r.ID, r.Author, r.Epithet)
	}
}

func decodeEntry(data []byte, start int) (Record, int) {
	var rec Record
	i := start

	// 1. Extract 7-character ID
	if i+7 <= len(data) {
		rec.ID = string(data[i : i+7])
		i += 7
	}

	var preName, mainName, epithet bytes.Buffer
	state := 0 // 0: pre-name/main-name, 1: inside name (after &1), 2: epithet

	for i < len(data) {
		// Only break if it's a confirmed new record start
		if i+4 < len(data) && isNewRecordStart(data[i:]) {
			break
		}

		b := data[i]

		// Termination on high-bit markers
		if b == 0xff || b == 0xfe || b == 0x83 {
			i++
			break
		}

		// Handle [2 and ]2 mapping
		if b == '[' && i+1 < len(data) && data[i+1] == '2' {
			writeToActiveBuffer(state, '(', &preName, &mainName, &epithet)
			i += 2
			continue
		}
		if b == ']' && i+1 < len(data) && data[i+1] == '2' {
			writeToActiveBuffer(state, ')', &preName, &mainName, &epithet)
			i += 2
			continue
		}

		// Handle &1 marker
		if b == '&' && i+1 < len(data) && data[i+1] == '1' {
			state = 1
			i += 2
			continue
		}

		// Handle closing & (Move to epithet)
		if b == '&' {
			// If we were in state 0 (no &1 found yet) or state 1, move to epithet
			state = 2
			i++
			continue
		}

		// Standard capture
		if b >= 32 && b < 127 {
			writeToActiveBuffer(state, b, &preName, &mainName, &epithet)
		}
		i++
	}

	// Assembly: If no &1 was found, the text is in preName
	mainStr := strings.TrimSpace(mainName.String())
	preStr := strings.TrimSpace(preName.String())

	if mainStr == "" && preStr != "" {
		// Entry had no &1, move preName to Author
		rec.Author = preStr
	} else {
		rec.Author = mainStr
		if preStr != "" {
			rec.Author = strings.Trim(preStr+" "+rec.Author, ", ")
		}
	}

	rec.Epithet = strings.TrimSpace(epithet.String())
	return rec, i
}

func writeToActiveBuffer(state int, char byte, pre, main, epi *bytes.Buffer) {
	switch state {
	case 0:
		pre.WriteByte(char)
	case 1:
		main.WriteByte(char)
	case 2:
		epi.WriteByte(char)
	}
}

func isNewRecordStart(buf []byte) bool {
	if len(buf) < 4 {
		return false
	}
	// Section markers like *CIV, *COP, *END
	if buf[0] == '*' {
		return true
	}

	prefix := string(buf[:3])
	// A valid author record MUST be Prefix + 4 Digits
	if prefix == "TLG" || prefix == "LAT" || prefix == "CIV" || prefix == "COP" {
		// Check if the next character is a digit
		if buf[3] >= '0' && buf[3] <= '9' {
			return true
		}
	}
	return false
}
