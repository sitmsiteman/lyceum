package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"tlgread/pkg/tlgcore"
)

type LemmaInfo struct {
	Lemma string
	Forms []string
}

func findForms(filePath, targetLemma string) (*LemmaInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		if len(parts) < 3 {
			continue
		}

		lemma := strings.TrimSpace(parts[0])
		if lemma == targetLemma {
			allForms := parts[2:]
			return &LemmaInfo{
				Lemma: lemma,
				Forms: allForms,
			}, nil
		}
	}

	return nil, fmt.Errorf("lemma %s not found", targetLemma)
}

func main() {

	fPath := flag.String("f", "greek-lemmata.txt", "file path for greek-lemmata.txt")
	word := flag.String("w", "", "word")
	isLatin := flag.Bool("l", false, "Search for latin words")
	flag.Parse()

	filePath := *fPath

	searchWord := *word
	for _, r := range *word {
		if r > 127 { // Simple check for non-ASCII
			searchWord = tlgcore.ToBetaCode(*word)
			break
		}
	}

	searchWord = tlgcore.NormalizeBetaCode(searchWord)

	info, err := findForms(filePath, searchWord)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !*isLatin {
		fmt.Printf("Lemma: %s\n", tlgcore.ToGreek(info.Lemma))
	} else {
		fmt.Printf("Lemma: %s\n", info.Lemma)
	}
	fmt.Println("Known inflections and variants:")
	for _, f := range info.Forms {
		if f != "" {
			form := strings.Split(f, " ")
			analysis := strings.Join(form[1:], " ")
			if !*isLatin {
				fmt.Printf(" - %s %s\n", tlgcore.ToGreek(form[0]), strings.TrimSpace(analysis))
			} else {
				fmt.Printf(" - %s %s\n", form[0], strings.TrimSpace(analysis))
			}
		}
	}
}
