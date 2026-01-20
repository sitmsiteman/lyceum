package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tlgread/pkg/tlgcore"
)

func main() {
	dirPath := flag.String("d", ".", "Directory containing TLG/PHI files")

	flag.Parse()

	fmt.Println("=== TLGRead-Go Feature Test Suite ===")
	fmt.Printf("Scanning directory: %s\n", *dirPath)

	// 1. Locate Author Table
	authPath := filepath.Join(*dirPath, "authtab.dir")
	hasAuth := false
	if _, err := os.Stat(authPath); err == nil {
		hasAuth = true
		fmt.Println("[PASS] Found authtab.dir")
	} else {
		fmt.Println("[WARN] authtab.dir not found. Author names will be 'Unknown'.")
	}

	// 2. Find all IDT files
	var idtFiles []string
	err := filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".idt") {
			idtFiles = append(idtFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("[FAIL] Error walking directory: %v\n", err)
		return
	}
	fmt.Printf("Found %d .idt files.\n", len(idtFiles))

	// 3. Select Files to Test
	// We want to test specific complex files if they exist, plus some randoms.
	priorityFiles := map[string]bool{
		"tlg0059": true, // Plato (Many works)
		"tlg0086": true, // Aristotle (Complex)
		"tlg0001": true, // Homer (Simple hierarchy)
		"tlg0012": true, // Iliad/Odyssey
		"phi0474": true, // Cicero (PHI format check)
	}

	var filesToTest []string
	var priorityFound []string

	// Separate priority from rest
	rest := []string{}
	for _, p := range idtFiles {
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
		if priorityFiles[base] {
			priorityFound = append(priorityFound, p)
		} else {
			rest = append(rest, p)
		}
	}

	// Pick up to 5 random files from the rest
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })

	filesToTest = append(filesToTest, priorityFound...)
	//	if len(rest) > 1820 {
	//		filesToTest = append(filesToTest, rest[:1820]...)
	//	} else {
	filesToTest = append(filesToTest, rest...)
	//	}

	fmt.Printf("Selected %d files for deep inspection.\n", len(filesToTest))
	fmt.Println("---------------------------------------------------")

	// 4. Run Tests
	passCount := 0
	failCount := 0

	for _, idtPath := range filesToTest {
		base := strings.TrimSuffix(filepath.Base(idtPath), filepath.Ext(idtPath))
		fmt.Printf("Testing %s ... ", base)

		// TEST A: Parse IDT
		meta, err := tlgcore.ReadIDT(idtPath)
		if err != nil {
			fmt.Printf("[FAIL] IDT Parse Error: %v\n", err)
			failCount++
			continue
		}
		if len(meta) == 0 {
			fmt.Printf("[FAIL] IDT Parsed but 0 works found (padding issue?)\n")
			failCount++
			continue
		}

		// TEST B: Check Author (if available)
		author := "N/A"
		if hasAuth {
			author = tlgcore.GetAuthorName(authPath, base)
			if author == "Unknown" || author == "" {
				// Not necessarily a fail, but worth noting
			}
		}

		// TEST C: Text Extraction
		txtPath := strings.TrimSuffix(idtPath, ".idt") + ".txt"
		if _, err := os.Stat(txtPath); os.IsNotExist(err) {
			// Try uppercase if linux
			txtPath = strings.TrimSuffix(idtPath, ".idt") + ".TXT"
		}

		if _, err := os.Stat(txtPath); err == nil {
			f, err := os.Open(txtPath)
			if err != nil {
				fmt.Printf("[FAIL] Could not open .txt: %v\n", err)
				failCount++
				continue
			}

			// Setup Parser
			p := tlgcore.NewParser(f)
			p.IDTData = meta
			if strings.HasPrefix(strings.ToLower(base), "lat") || strings.HasPrefix(strings.ToLower(base), "phi") {
				p.IsLatinFile = true
			}

			// Extract first work ID
			var firstWorkID string
			for id := range meta {
				firstWorkID = id
				break // Just take one
			}

			text, err := p.ExtractWork(firstWorkID)
			f.Close()

			if err != nil {
				fmt.Printf("[FAIL] ExtractWork(%s) error: %v\n", firstWorkID, err)
				failCount++
				continue
			}

			if len(text) < 10 {
				fmt.Printf("[FAIL] Extracted text too short (<10 chars). Empty work?\n")
				failCount++
				continue
			}

			// Success
			fmt.Printf("[PASS] Works: %d | Author: %.15s... | Text extracted: %d bytes\n", len(meta), author, len(text))
			passCount++

		} else {
			fmt.Printf("[WARN] .txt file missing, skipped text check.\n")
			passCount++ // Counting as pass for IDT logic
		}
	}

	fmt.Println("---------------------------------------------------")
	fmt.Printf("Test Complete. Passed: %d, Failed: %d\n", passCount, failCount)
}
