package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tlgread/pkg/tlgcore"
)

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
	idtData, err := tlgcore.ReadIDT(idtPath)

	if err != nil {
		fmt.Printf("Warning: Failed to read IDT file %s: %v\n", idtPath, err)
		idtData = make(map[string]*tlgcore.WorkMetadata)
	}

	authPath := filepath.Join(dir, "authtab.dir")
	var author string = "Unknown Author"

	records, err := tlgcore.ReadAuthorTable(authPath)
	if err == nil {
		targetID := strings.ToUpper(tlgID)
		for _, rec := range records {
			if strings.TrimSpace(rec.ID) == targetID {
				author = rec.Name
				break
			}
		}
	} else {
		fmt.Printf("Warning: Could not read author table: %v\n", err)
	}

	p := tlgcore.NewParser(f)
	p.IDTData = idtData

	latinBase := []string{"LAT", "CIV", "PHI"}

	for _, pref := range latinBase {
		if strings.HasPrefix(strings.ToUpper(base), pref) {
			p.IsLatinFile = true
			break
		}
	}

	if *list {
		fmt.Printf("File: %s (%s)\n", base, author)
		fmt.Println("----------------------------------------")

		works, err := p.ExtractList(idtData)
		if err != nil {
			log.Fatal(err)
		}
		for _, w := range works {
			fmt.Println(w)
		}

	} else {
		cleanWID := tlgcore.NormalizeID(*wID)

		title := "(Unknown Title)"
		meta := idtData[cleanWID]
		if meta != nil {
			title = meta.Title
		}

		fmt.Printf("Author: %s\nWork:   %s (ID: %s)\n", author, title, cleanWID)

		if meta != nil && len(meta.Citations) > 0 {
			for _, c := range meta.Citations {
				fmt.Printf("%s (%s) ", c.Label, c.LevelChar)
			}
			fmt.Printf("\n")
		}
		fmt.Println("----------------------------------------")

		text, err := p.ExtractWork(cleanWID)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Print(text)
		}
	}
}
