package main

import (
	"fmt"
	"strconv"
	"strings"
)

// FretDef represents a fret with an optional note name.
type FretDef struct {
	Fret  int
	Note  string // optional note name (e.g., "R", "3", "5")
	Muted bool   // true if the string is not played ("x")
}

// Chord represents a ChordPro chord definition.
type Chord struct {
	Name      string
	StartFret int
	Frets     []FretDef
}

// ParseChord parses a ChordPro chord definition string like
// "{define: E 0 2 2 1 0 0}" into a Chord.
// The format is: {define: <name> <start_fret> <fret.note> <fret.note> <fret.note> <fret.note> <fret.note> <fret.note>}
// where <start_fret> is the starting fret and exactly 6 fret definitions follow.
func ParseChord(input string) (Chord, error) {
	s := strings.TrimSpace(input)
	s, _ = strings.CutPrefix(s, "{")
	s, _ = strings.CutSuffix(s, "}")
	s = strings.TrimSpace(s)

	prefix, rest, found := strings.Cut(s, ":")
	if !found || strings.TrimSpace(prefix) != "define" {
		return Chord{}, fmt.Errorf("expected '{define: ...}', got %q", input)
	}

	fields := strings.Fields(rest)
	// Need: name + start_fret + 6 fret defs = 8 fields
	if len(fields) < 8 {
		return Chord{}, fmt.Errorf("not enough fields in chord definition (need name, start fret, and 6 fret values): %q", input)
	}

	name := fields[0]

	startFret, err := strconv.Atoi(fields[1])
	if err != nil {
		return Chord{}, fmt.Errorf("invalid start fret %q: %w", fields[1], err)
	}

	frets := make([]FretDef, 0, 6)
	for _, f := range fields[2:8] {
		fretPart, note, _ := strings.Cut(f, ".")
		if strings.EqualFold(fretPart, "x") {
			frets = append(frets, FretDef{Muted: true})
			continue
		}
		n, err := strconv.Atoi(fretPart)
		if err != nil {
			return Chord{}, fmt.Errorf("invalid fret value %q: %w", f, err)
		}
		frets = append(frets, FretDef{Fret: n, Note: note})
	}

	if len(frets) != 6 {
		return Chord{}, fmt.Errorf("expected exactly 6 fret definitions, got %d", len(frets))
	}

	return Chord{Name: name, StartFret: startFret, Frets: frets}, nil
}
