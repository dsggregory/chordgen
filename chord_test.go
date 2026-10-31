package main

import (
	"slices"
	"testing"
)

func TestParseChord(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantName      string
		wantStartFret int
		wantFrets     []FretDef
		wantErr       bool
	}{
		{
			name:          "standard E chord",
			input:         "{define: E 0 0 2 2 1 0 0}",
			wantName:      "E",
			wantStartFret: 0,
			wantFrets:     []FretDef{{0, "", false}, {2, "", false}, {2, "", false}, {1, "", false}, {0, "", false}, {0, "", false}},
		},
		{
			name:          "Am chord",
			input:         "{define: Am 0 x 0 2 2 1 0}",
			wantName:      "Am",
			wantStartFret: 0,
			wantFrets:     []FretDef{{Muted: true}, {0, "", false}, {2, "", false}, {2, "", false}, {1, "", false}, {0, "", false}},
		},
		{
			name:          "E chord with note names",
			input:         "{define: E 0 0.R 2.3 2.5 1 0 0}",
			wantName:      "E",
			wantStartFret: 0,
			wantFrets:     []FretDef{{0, "R", false}, {2, "3", false}, {2, "5", false}, {1, "", false}, {0, "", false}, {0, "", false}},
		},
		{
			name:          "barre chord with start fret",
			input:         "{define: F 1 1 1 2 3 3 1}",
			wantName:      "F",
			wantStartFret: 1,
			wantFrets:     []FretDef{{1, "", false}, {1, "", false}, {2, "", false}, {3, "", false}, {3, "", false}, {1, "", false}},
		},
		{
			name:    "not enough fret values",
			input:   "{define: E 0 2 2 1 0}",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid fret value",
			input:   "{define: E 0 abc 2 1 0 0 0}",
			wantErr: true,
		},
		{
			name:          "muted strings",
			input:         "{define: D 0 x x 0 2 3 2}",
			wantName:      "D",
			wantStartFret: 0,
			wantFrets:     []FretDef{{Muted: true}, {Muted: true}, {0, "", false}, {2, "", false}, {3, "", false}, {2, "", false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chord, err := ParseChord(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if chord.Name != tt.wantName {
				t.Errorf("name = %q, want %q", chord.Name, tt.wantName)
			}
			if chord.StartFret != tt.wantStartFret {
				t.Errorf("startFret = %d, want %d", chord.StartFret, tt.wantStartFret)
			}
			if !slices.Equal(chord.Frets, tt.wantFrets) {
				t.Errorf("frets = %v, want %v", chord.Frets, tt.wantFrets)
			}
		})
	}
}
