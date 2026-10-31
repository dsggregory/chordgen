package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// parseTitle checks if a line is a {title: ...} directive and returns the title.
func parseTitle(line string) (string, bool) {
	s := strings.TrimSpace(line)
	s, _ = strings.CutPrefix(s, "{")
	s, _ = strings.CutSuffix(s, "}")
	s = strings.TrimSpace(s)
	key, val, found := strings.Cut(s, ":")
	if !found || strings.TrimSpace(strings.ToLower(key)) != "title" {
		return "", false
	}
	return strings.TrimSpace(val), true
}

func formatFrets(frets []FretDef) string {
	parts := make([]string, len(frets))
	for i, f := range frets {
		if f.Muted {
			parts[i] = "x"
		} else if f.Note != "" {
			parts[i] = fmt.Sprintf("%d.%s", f.Fret, f.Note)
		} else {
			parts[i] = fmt.Sprintf("%d", f.Fret)
		}
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func processChord(line string) {
	chord, err := ParseChord(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Chord: %s, StartFret: %d, Frets: %v\n", chord.Name, chord.StartFret, formatFrets(chord.Frets))

	outFile := chord.Name + ".pdf"
	if err := DrawChordDiagram(chord, outFile, line); err != nil {
		fmt.Fprintf(os.Stderr, "error generating diagram: %v\n", err)
		return
	}
	fmt.Printf("Wrote %s\n", outFile)
}

func main() {
	if len(os.Args) > 1 {
		// Treat all command-line arguments as a single chord definition
		line := strings.Join(os.Args[1:], " ")
		chord, err := ParseChord(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Chord: %s, StartFret: %d, Frets: %v\n", chord.Name, chord.StartFret, formatFrets(chord.Frets))
		outFile := chord.Name + ".pdf"
		if err := DrawChordDiagram(chord, outFile, line); err != nil {
			fmt.Fprintf(os.Stderr, "error generating diagram: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", outFile)
		return
	}

	// Read chord definitions from stdin and collect them
	var chords []Chord
	var title string
	var inputLines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		inputLines = append(inputLines, line)
		if t, ok := parseTitle(line); ok {
			title = t
			continue
		}
		chord, err := ParseChord(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Printf("Chord: %s, StartFret: %d, Frets: %v\n", chord.Name, chord.StartFret, formatFrets(chord.Frets))
		chords = append(chords, chord)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	if len(chords) == 0 {
		return
	}

	sourceData := strings.Join(inputLines, "\n")

	// Single chord: generate individual file; multiple: generate a page
	if len(chords) == 1 {
		outFile := chords[0].Name + ".pdf"
		if err := DrawChordDiagram(chords[0], outFile, sourceData); err != nil {
			fmt.Fprintf(os.Stderr, "error generating diagram: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", outFile)
	} else {
		outFile := "chords.pdf"
		if title != "" {
			outFile = title + ".pdf"
		}
		if err := DrawChordPage(chords, title, outFile, sourceData); err != nil {
			fmt.Fprintf(os.Stderr, "error generating page: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote \"%s\"\n", outFile)
	}
}
