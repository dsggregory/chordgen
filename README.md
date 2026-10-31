# chordgen

A command-line tool that parses ChordPro chord definitions and generates PDF chord diagrams.

## Usage

### Single chord from command line

```bash
chordgen '{define: E 0 0 2 2 1 0 0}'
```

### Multiple chords from stdin

In the resulting PDF, you can grep for the "/Subject" string to find the raw chord definitions used.

```bash
cat chords.txt | chordgen
```

Or interactively:

```bash
chordgen
{define: E 0 0 2 2 1 0 0}
{define: Am 0 x 0 2 2 1 0}
^D
```

### Output name
If a `{title: My Title}` directive is present in the input, the output file is based on the title.

Otherwise, a single chord definition produces a `<ChordName>.pdf` diagram file in the current directory.
A multi-chord produces a `chords.pdf` file in the current directory
## Input Syntax

```
{define: <name> <start_fret> <f1> <f2> <f3> <f4> <f5> <f6>}
```

| Field          | Description                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `<name>`       | Chord name (e.g., `E`, `Am`, `F#m7`)                                       |
| `<start_fret>` | Integer indicating the starting fret (offset from the nut). `0` = open position. |
| `<f1>`–`<f6>` | Exactly 6 fret definitions, one per string (low E to high E), each with the format described below. |

### Fret Definition Format

Each of the 6 fret values uses the format `<offset>` or `<offset>.<note_name>`:

- **`<offset>`** — Integer offset from the start fret where the finger is placed.
  - `0` — String is played open. Displayed as **O** above the nut on the diagram.
- **`x`** — String is muted (not played). Displayed as **X** above the nut on the diagram.
- **`.<note_name>`** *(optional)* — A period followed by a note name (e.g., `.R` for root, `.b3` for minor third, `.5` for fifth).

### Examples

| Definition                              | Description                                      |
|-----------------------------------------|--------------------------------------------------|
| `{define: E 0 0 2 2 1 0 0}`            | Open E chord at the nut                          |
| `{define: Am 0 x 0 2 2 1 0}`           | Am chord with muted low E string                 |
| `{define: Em 0 0.R 2.5 2 0.b3 0 0}`      | Em chord with note names (Root, flat-3rd, 5th)         |
| `{define: F 1 1 1 2 3 3 1}`            | F barre chord starting at fret 1                 |
| `{define: D 0 x x 0 2 3 2}`            | D chord with two muted strings                   |

## Output

For each parsed chord, the tool:

1. Prints the chord details to stdout.
2. Generates a PDF chord diagram showing:
   - A 5×5 grid with vertical lines (strings) and horizontal lines (frets).
   - A thick top line representing the nut.
   - **X** above muted strings and **O** above open strings.
   - Filled dots at fretted positions.
   - The chord name and start fret label.
