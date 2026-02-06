package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"tlgread/pkg/tlgcore"
)

func main() {

	xPath := flag.String("f", "grc.lsj.xml", "file path for dictionary xml file")
	iPath := flag.String("o", "lsj.idt", "file path for export index file")
	flag.Parse()

	xmlPath := *xPath
	indexPath := *iPath

	f, err := os.Open(xmlPath)
	if err != nil {
		fmt.Println("Error: Cannot find grc.lsj.xml")
		return
	}
	defer f.Close()

	out, _ := os.Create(indexPath)
	defer out.Close()

	reader := bufio.NewReader(f)
	var offset int64
	re := regexp.MustCompile(`key="([^"]+)"`)

	fmt.Println("Indexing", xmlPath, "... this may take a few seconds.")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.HasPrefix(line, "<div2") {
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				rawKey := match[1]
				strictKey := tlgcore.NormalizeStrict(rawKey)
				fuzzyKey := tlgcore.NormalizeFuzzy(rawKey)

				fmt.Fprintf(out, "'%s' => %d\n", strictKey, offset)
				if fuzzyKey != strictKey {
					fmt.Fprintf(out, "'%s' => %d\n", fuzzyKey, offset)
				}
			}
		}

		if strings.HasPrefix(line, "<div1") {
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				rawKey := match[1]
				strictKey := tlgcore.NormalizeLatin(rawKey)

				fmt.Fprintf(out, "'%s' => %d\n", strictKey, offset)
			}
		}
		offset += int64(len(line))
	}
	fmt.Println("Done!", indexPath, "created.")
}
