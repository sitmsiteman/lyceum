package tlgcore

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CanonField struct {
	Tag   string
	Label string
	Value string
}

var canonTagMap = map[string]string{
	"nam": "Author Name", "epi": "Epithet", "geo": "Geography", "dat": "Date",
	"vid": "Vide", "wrk": "Work Title", "cla": "Classification", "xmt": "Transmission",
	"typ": "Type", "wct": "Word Count", "cit": "Citation Schema", "tit": "Title in Ed.",
	"pub": "Publisher", "pla": "Place", "pyr": "Pub. Year", "ryr": "Reprint Year",
	"pag": "Pages", "edr": "Editor", "brk": "Breaks/Frags", "ser": "Series", "key": "Key ID",
}

func GetBiblioFromCanon(canonPath string, tlgID string, workID string) (string, error) {
	f, err := os.Open(canonPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	p := NewParser(f)
	fullText, err := p.ExtractAllText()
	if err != nil {
		return "", fmt.Errorf("canon parse error: %v", err)
	}

	cleanID := strings.ToUpper(tlgID)
	cleanID = strings.TrimPrefix(cleanID, "TLG")
	cleanID = strings.TrimLeft(cleanID, "0")
	if cleanID == "" {
		cleanID = "0"
	}

	idNum, _ := strconv.Atoi(cleanID)
	authID := fmt.Sprintf("%04d", idNum)

	var wID string
	if workID != "" {
		cleanWID := strings.TrimLeft(workID, "0")
		wIDNum, _ := strconv.Atoi(cleanWID)
		wID = fmt.Sprintf("%03d", wIDNum)
	}

	scanner := bufio.NewScanner(strings.NewReader(fullText))

	var authorBuffer strings.Builder
	var workBuffer strings.Builder

	inAuthorSection := false
	inWorkSection := false

	targetWorkPattern := fmt.Sprintf("%s %s", authID, wID)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		isIDLine := len(line) >= 4 && isNumeric(line[:4])

		if isIDLine {
			if strings.HasPrefix(line, authID) {
				isWorkLine := len(line) >= 8 && line[4] == ' ' && isNumeric(line[5:8])

				if !isWorkLine {
					inAuthorSection = true
					inWorkSection = false
					authorBuffer.WriteString(line + "\n")
					continue
				} else {
					if wID != "" && strings.HasPrefix(line, targetWorkPattern) {
						inWorkSection = true
						workBuffer.WriteString("\n" + line + "\n")
					} else {
						inWorkSection = false
					}
				}
			} else {
				if inAuthorSection {
					break
				}
			}
		} else {
			if inWorkSection {
				workBuffer.WriteString(line + "\n")
			} else if inAuthorSection {
				authorBuffer.WriteString(line + "\n")
			}
		}
	}

	if workBuffer.Len() > 0 {
		return strings.TrimSpace(workBuffer.String()), nil
	}
	if authorBuffer.Len() > 0 {
		return strings.TrimSpace(authorBuffer.String()), nil
	}

	return "", nil
}

func GetMetadataFromCanonDB(dbPath string, tlgID string, workID string) ([]CanonField, error) {
	f, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := NewParser(f)
	fullText, err := p.ExtractAllText()
	if err != nil {
		return nil, err
	}

	cleanID := strings.ToUpper(tlgID)
	cleanID = strings.TrimPrefix(cleanID, "TLG")
	cleanID = strings.TrimLeft(cleanID, "0")
	if cleanID == "" {
		cleanID = "0"
	}
	idNum, _ := strconv.Atoi(cleanID)
	authID := fmt.Sprintf("%04d", idNum)

	authKey := fmt.Sprintf("key %s", authID)
	targetWorkKey := ""
	if workID != "" {
		cleanWID := strings.TrimLeft(workID, "0")
		wIDNum, _ := strconv.Atoi(cleanWID)
		wID := fmt.Sprintf("%03d", wIDNum)
		targetWorkKey = fmt.Sprintf("key %s %s", authID, wID)
	}

	var fields []CanonField
	scanner := bufio.NewScanner(strings.NewReader(fullText))

	capture := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "key ") {
			isAuthMatch := (line == authKey)
			isWorkMatch := (targetWorkKey != "" && line == targetWorkKey)

			if isAuthMatch || isWorkMatch {
				capture = true
				sectionName := "Author Metadata"
				if isWorkMatch {
					sectionName = "Work Metadata"
				}
				fields = append(fields, CanonField{Tag: "---", Label: "Section", Value: sectionName})
			} else {
				capture = false
			}
		}

		if capture {
			if len(line) > 4 && line[3] == ' ' {
				tag := line[:3]
				if tag == "key" {
					continue
				}

				val := strings.TrimSpace(line[4:])
				label, exists := canonTagMap[tag]
				if !exists {
					label = tag
				}
				fields = append(fields, CanonField{Tag: tag, Label: label, Value: val})
			}
		}
	}
	return fields, nil
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
