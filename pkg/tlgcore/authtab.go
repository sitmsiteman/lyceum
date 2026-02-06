package tlgcore

import (
	"os"
	"strings"
)

type AuthorRecord struct {
	ID   string
	Name string
}

func ReadAuthorTable(path string) ([]AuthorRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var records []AuthorRecord
	i := 0
	for i < len(data) {
		if isNewRecordStart(data[i:]) {
			rec, nextPos := decodeAuthorEntry(data, i)
			records = append(records, rec)
			i = nextPos
			continue
		}
		i++
	}

	return records, nil
}

func decodeAuthorEntry(data []byte, start int) (AuthorRecord, int) {
	var rec AuthorRecord
	i := start

	if i+8 <= len(data) {
		rec.ID = strings.TrimSpace(string(data[i : i+8]))
		i += 8
	}

	var textParts []string
	currentFieldType := 0

	for i < len(data) {
		if i+4 < len(data) && isNewRecordStart(data[i:]) {
			break
		}

		b := data[i]

		if b == 0xFF {
			i++
			for i < len(data) && data[i] == 0xFF {
				i++
			}
			break
		}

		if b >= 0x80 {
			currentFieldType = int(b)
			i++
			continue
		}

		startText := i
		for i < len(data) {
			if data[i] >= 0x80 {
				break
			}
			if i+4 < len(data) && isNewRecordStart(data[i:]) {
				break
			}
			i++
		}

		segment := data[startText:i]
		if len(segment) > 0 {
			if currentFieldType == 0x83 {
				continue
			}

			decoded := ToLatin(string(segment))
			trimmed := strings.TrimSpace(decoded)
			if trimmed != "" {
				textParts = append(textParts, trimmed)
			}
		}
	}

	rec.Name = strings.Join(textParts, " ")
	return rec, i
}

func isNewRecordStart(buf []byte) bool {
	if len(buf) < 4 {
		return false
	}

	prefix := string(buf[:3])
	if prefix == "TLG" || prefix == "LAT" || prefix == "CIV" || prefix == "COP" || prefix == "L  " {
		if buf[3] >= '0' && buf[3] <= '9' {
			return true
		}
	}
	return false
}
