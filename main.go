package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

var monthMap = map[time.Month]string{
	time.January:   "Jänner",
	time.February:  "Februar",
	time.March:     "März",
	time.April:     "April",
	time.May:       "Mai",
	time.June:      "Juni",
	time.July:      "Juli",
	time.August:    "August",
	time.September: "September",
	time.October:   "Oktober",
	time.November:  "November",
	time.December:  "Dezember",
}

func main() {
	// Kommandozeilenparameter
	sourceDir := flag.String("source", "", "Quellverzeichnis mit Mediendateien (erforderlich)")
	destDir := flag.String("destination", "", "Zielverzeichnis für sortierte Dateien (optional, Standard: gleich wie --source)")
	dryRun := flag.Bool("dry-run", false, "Nur anzeigen, was verschoben würde, ohne es zu tun")
	flag.Parse()

	// Validierung
	if *sourceDir == "" {
		fmt.Println("❌ Fehler: --source muss angegeben werden")
		flag.Usage()
		os.Exit(1)
	}

	// Wenn kein Zielverzeichnis angegeben ist, verwende das Quellverzeichnis
	finalDestDir := *destDir
	if finalDestDir == "" {
		finalDestDir = *sourceDir
	}

	// Durchsuche alle Dateien im Quellverzeichnis (ohne Unterordner)
	err := filepath.WalkDir(*sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Lesen:", err)
			return nil
		}

		// Nur Dateien im angegebenen Verzeichnis (keine Unterordner)
		if d.IsDir() || filepath.Dir(path) != *sourceDir {
			return nil
		}

		// Datei öffnen
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Kann Datei nicht öffnen:", path)
			return nil
		}
		defer file.Close()

		// Versuche Exif-Datum zu lesen
		var taken time.Time
		x, err := exif.Decode(file)
		if err == nil {
			tm, err := x.DateTime()
			if err == nil {
				taken = tm
			}
		}

		// Fallback auf Dateisystemzeit
		if taken.IsZero() {
			info, err := file.Stat()
			if err != nil {
				fmt.Println("Kann Stat nicht lesen:", path)
				return nil
			}
			taken = info.ModTime()
		}

		year := taken.Year()
		month := monthMap[taken.Month()]
		targetDir := filepath.Join(finalDestDir, fmt.Sprintf("%d", year), month)
		destPath := filepath.Join(targetDir, filepath.Base(path))

		if *dryRun {
			fmt.Printf("[Dry Run] %s → %s\n", filepath.Base(path), destPath)
			return nil
		}

		// Zielverzeichnis erstellen
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			fmt.Println("Fehler beim Erstellen von:", targetDir)
			return nil
		}

		// Datei verschieben
		err = os.Rename(path, destPath)
		if err != nil {
			fmt.Println("Fehler beim Verschieben:", err)
		} else {
			fmt.Printf("Verschoben: %s → %d/%s\n", filepath.Base(path), year, month)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Fehler beim Verarbeiten:", err)
	}
}