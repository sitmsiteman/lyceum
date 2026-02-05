package main

import (
	"flag"
	"fmt"
	"log"

	"tlgread/pkg/tlgcore"
)

func main() {
	fPath := flag.String("f", "authtab.dir", "filename")
	flag.Parse()

	records, err := tlgcore.ReadAuthorTable(*fPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range records {
		if len(r.ID) > 0 && r.ID[0] != '*' {
			fmt.Printf("%-8s | %s\n", r.ID, r.Name)
		}
	}
}
