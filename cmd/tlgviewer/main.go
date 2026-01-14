package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tlgread/pkg/tlgcore"
)

func getTitlesFromIDT(path string) map[string]string {
	m := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}

	idCounter := 1
	for i := 0; i < len(data)-3; i++ {
		// Signature: 10 01 (Description for Work Level)
		if data[i] == 0x10 && data[i+1] == 0x01 {
			length := int(data[i+2])
			if length == 0 || length > 200 {
				continue
			}
			if i+3+length > len(data) {
				continue
			}

			rawTitle := string(data[i+3 : i+3+length])

			// Keep Latin titles as is, convert only if Beta Code (contains *)
			cleanTitle := rawTitle
			if strings.Contains(rawTitle, "*") {
				cleanTitle = tlgcore.ToGreek(rawTitle)
			}

			m[strconv.Itoa(idCounter)] = cleanTitle
			idCounter++

			i += 2 + length
		}
	}
	return m
}

func getAuthorName(path, tlgID string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "Unknown"
	}
	cleanID := fmt.Sprintf("TLG%04s", strings.TrimPrefix(strings.ToUpper(tlgID), "TLG"))
	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?&1(.*?)&`, cleanID))
	m := re.FindSubmatch(data)
	if len(m) > 1 {
		return strings.TrimSpace(strings.Split(string(m[1]), "&")[0])
	}
	return tlgID
}

func main() {
	fPath := flag.String("f", "", "TLG .txt")
	wID := flag.String("w", "", "Work ID")
	list := flag.Bool("list", false, "List")
	flag.Parse()

	if *fPath == "" {
		log.Fatal("Usage: ./tlgviewer -f tlg[0000-9999].txt [-list] or [-w 1]")
	}

	f, err := os.Open(*fPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dir, base := filepath.Split(*fPath)
	tlgID := strings.TrimSuffix(base, filepath.Ext(base))

	idtPath := filepath.Join(dir, tlgID+".idt")
	titles := getTitlesFromIDT(idtPath)

	authPath := filepath.Join(dir, "authtab.dir")
	author := getAuthorName(authPath, tlgID)

	if *list {
		fmt.Printf("File: %s (%s)\n", base, author)
		fmt.Println("----------------------------------------")
	} else {
		t := titles[*wID]
		if t == "" {
			t = titles[tlgcore.NormalizeID(*wID)]
		}
		if t == "" {
			t = "(Unknown Title)"
		}
		fmt.Printf("Author: %s\nWork:   %s (ID: %s)\n", author, t, *wID)
		fmt.Println("----------------------------------------")
	}

	p := tlgcore.NewParser(f)
	p.Run(*wID, *list, titles)
}
