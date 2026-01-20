package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"tlgread/pkg/tlgcore"
)

type MorphResult struct {
	Form       string
	Lemma      string
	ShortDef   string
	Morphology string
}

type LSJEntry struct {
	Key   string `xml:"key,attr"`
	Orth  string `xml:"orth"`
	Sense string `xml:",innerxml"`
}

func LoadIndex(idtPath string) (map[string]int64, []string, error) {
	index := make(map[string]int64)
	var keys []string
	file, err := os.Open(idtPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	re := regexp.MustCompile(`'(.+?)' => (\d+)`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) == 3 {
			offset, _ := strconv.ParseInt(matches[2], 10, 64)
			key := matches[1]
			index[key] = offset
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return index, keys, nil
}

func FindLemmaIndexed(filePath string, offset int64, searchForm string) ([]MorphResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	file.Seek(offset, 0)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	re := regexp.MustCompile(`\{[^ ]+ \d+ (?:[^,]+,)?(?P<lemma>[^ ]+)(?P<content>.*?)\}`)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		currentWord := strings.TrimPrefix(fields[0], "!")

		if strings.EqualFold(currentWord, searchForm) {
			var results []MorphResult
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				parts := regexp.MustCompile(`\s{2,}`).Split(strings.TrimSpace(match[2]), -1)

				resDef := "---"
				resMorph := ""
				if len(parts) >= 2 {
					resDef = strings.TrimSpace(parts[0])
					resMorph = strings.TrimSpace(parts[1])
				} else if len(parts) == 1 {
					resMorph = strings.TrimSpace(parts[0])
				}

				results = append(results, MorphResult{
					Form:       searchForm,
					Lemma:      strings.TrimSpace(match[1]), // Final fix for trailing spaces
					ShortDef:   resDef,
					Morphology: resMorph,
				})
			}
			return results, nil
		}
		if len(currentWord) > 0 && currentWord[0] > searchForm[0] {
			break
		}
	}
	return nil, fmt.Errorf("not found")
}

func lookupLSJ(xmlPath string, rawLemma string, lsjIndex map[string]int64, seenOffsets map[int64]bool, isLSJ bool) {
	var strictKey string

	lemma := strings.Fields(rawLemma)[0]

	if isLSJ {
		strictKey = tlgcore.NormalizeStrict(lemma)
	} else {
		strictKey = tlgcore.NormalizeLatin(lemma)
	}

	fuzzyKey := tlgcore.NormalizeFuzzy(lemma)

	var offsets []int64
	localSeen := make(map[int64]bool)

	addUnique := func(val int64) {
		if !localSeen[val] {
			offsets = append(offsets, val)
			localSeen[val] = true
		}
	}

	// Check the base key (e.g., "legw")
	if val, ok := lsjIndex[strictKey]; ok {
		addUnique(val)
	}

	// Check numbered keys (e.g., "legw2", "legw3", ...) ... Do we need this?
	for i := 2; ; i++ {
		key := strictKey + strconv.Itoa(i)
		val, ok := lsjIndex[key]
		if !ok {
			break
		}
		addUnique(val)
	}

	if len(offsets) == 0 {
		if val, ok := lsjIndex[fuzzyKey]; ok {
			addUnique(val)
		} else {
			// Prefix scan fallback
			for k, off := range lsjIndex {
				if strings.HasPrefix(k, fuzzyKey) {
					addUnique(off)
					break // Stop after first fuzzy match
				}
			}
		}
	}

	if len(offsets) > 0 {
		f, err := os.Open(xmlPath)
		if err != nil {
			fmt.Println("Error opening LSJ file:", err)
			return
		}
		defer f.Close()

		for _, offset := range offsets {
			if seenOffsets[offset] {
				continue
			}

			_, err = f.Seek(offset, 0)
			if err != nil {
				fmt.Println("Seek error:", err)
				return
			}

			decoder := xml.NewDecoder(f)
			var entry LSJEntry
			err = decoder.Decode(&entry)
			if err != nil {
				return
			}

			seenOffsets[offset] = true

			if isLSJ {
				fmt.Printf("\n[ENTRY: %s]\n", tlgcore.ToGreek(entry.Key))
			} else {
				fmt.Printf("\n[ENTRY: %s]\n", entry.Key)
			}
			fmt.Printf("%s\n", processSense(entry.Sense))
		}
	}
}

func LoadLSJIndex(path string) map[string]int64 {
	index := make(map[string]int64)
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Warning: Could not open index file at %s\n", path)
		return index
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Split by ' => ' which is what our indexer uses
		parts := strings.Split(line, " => ")
		if len(parts) == 2 {
			// Clean the key: remove surrounding single quotes
			key := strings.Trim(parts[0], "'")
			offset, _ := strconv.ParseInt(parts[1], 10, 64)
			index[key] = offset
		}
	}
	return index
}

func processSense(rawXml string) string {
	// 1. Convert Greek tags to Unicode Greek first
	reForeign := regexp.MustCompile(`<foreign lang="greek">([^<]+)</foreign>`)
	processed := reForeign.ReplaceAllStringFunc(rawXml, func(match string) string {
		code := reForeign.FindStringSubmatch(match)[1]
		return tlgcore.ToGreek(code)
	})

	// 2. Structural Replacements: Turn tags into layout markers
	// Treat each sense as a new paragraph with a newline
	processed = strings.ReplaceAll(processed, "<sense", "\n\n  • <sense")

	// Ensure bibliographic references have a space after them
	processed = strings.ReplaceAll(processed, "</bibl>", " ")
	processed = strings.ReplaceAll(processed, "</cit>", " ")

	// 3. Strip all remaining XML tags
	stripTags := regexp.MustCompile("<[^>]*>")
	clean := stripTags.ReplaceAllString(processed, "")

	// 4. Decode XML entities
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&quot;", "\"")

	// 5. Clean up horizontal whitespace
	// We preserve the double newlines we created, but collapse extra spaces on lines
	lines := strings.Split(clean, "\n")
	var finalLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Collapse multiple spaces within the line
			reSpace := regexp.MustCompile(`\s+`)
			finalLines = append(finalLines, reSpace.ReplaceAllString(trimmed, " "))
		}
	}

	return strings.Join(finalLines, "\n\n")
}

func main() {
	wordRaw := flag.String("w", "", "word in Beta Code / Greek")
	lsjPath := flag.String("dic", "grc.lsj.xml", "LSJ XML path")
	idtPath := flag.String("idt", "greek-analyses.idt", "idt file")
	analPath := flag.String("a", "greek-analyses.txt", "analyses txt file")
	lsjidtPath := flag.String("dicidt", "lsj.idt", "LSJ idt file")
	printdic := flag.Bool("entry", true, "print dictionary entries or not")
	isLatin := flag.Bool("lat", false, "use L-S dictionary")

	flag.Parse()

	lsjIndex := LoadLSJIndex(*lsjidtPath)

	searchWord := *wordRaw
	for _, r := range *wordRaw {
		if r > 127 { // Simple check for non-ASCII
			searchWord = tlgcore.ToBetaCode(*wordRaw)
			break
		}
	}

	searchWord = tlgcore.NormalizeBetaCode(searchWord)

	index, keys, _ := LoadIndex(*idtPath)
	idx := sort.SearchStrings(keys, searchWord)
	if idx > 0 {
		idx -= 1
	}

	var results []MorphResult
	var err error
	for i := range 3 {
		if idx-i < 0 {
			break
		}
		results, err = FindLemmaIndexed(*analPath, index[keys[idx-i]], searchWord)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatal("Morphology not found.")
	}

	// Output Morph and then LSJ

	seenLSJEntries := make(map[int64]bool)

	for _, r := range results {
		// 1. Print Morphology
		lemmaDisplay := strings.Fields(r.Lemma)[0]
		if !*isLatin {
			fmt.Printf("Greek: %s | Lemma: %s (%s)\n", tlgcore.ToGreek(searchWord), tlgcore.ToGreek(lemmaDisplay), r.Morphology)
		} else {
			fmt.Printf("Latin: %s | Lemma: %s (%s)\n", searchWord, lemmaDisplay, r.Morphology)
		}
	}
	if *printdic == true {
		for _, r := range results {
			lookupLSJ(*lsjPath, r.Lemma, lsjIndex, seenLSJEntries, !*isLatin)
		}
	}
}
