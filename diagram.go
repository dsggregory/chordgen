package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	gridCols   = 5 // number of strings (vertical lines)
	gridRows   = 5 // number of frets (horizontal lines)
	cellWidth  = 40
	cellHeight = 40
	margin     = 40
	topMargin  = 60 // extra space above nut for X/O markers and chord name
	dotRadius  = 8
)

// chordWidth returns the width of a single chord diagram.
func chordWidth(numStrings int) int {
	return margin*2 + (numStrings-1)*cellWidth
}

// chordHeight returns the height of a single chord diagram.
func chordHeight() int {
	return topMargin + margin + (gridRows-1)*cellHeight + margin
}

// pdfBuilder accumulates PDF drawing commands in a content stream.
type pdfBuilder struct {
	stream strings.Builder
}

func (p *pdfBuilder) setColor(r, g, b float64) {
	p.stream.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg %.3f %.3f %.3f RG\n", r, g, b, r, g, b))
}

func (p *pdfBuilder) setFillColor(r, g, b float64) {
	p.stream.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg\n", r, g, b))
}

func (p *pdfBuilder) setStrokeColor(r, g, b float64) {
	p.stream.WriteString(fmt.Sprintf("%.3f %.3f %.3f RG\n", r, g, b))
}

func (p *pdfBuilder) setLineWidth(w float64) {
	p.stream.WriteString(fmt.Sprintf("%.2f w\n", w))
}

func (p *pdfBuilder) line(x1, y1, x2, y2 float64) {
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n", x1, y1, x2, y2))
}

func (p *pdfBuilder) filledCircle(cx, cy, r float64) {
	// Approximate circle with 4 Bezier curves
	k := r * 0.5523
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f m\n", cx, cy+r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+k, cy+r, cx+r, cy+k, cx+r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+r, cy-k, cx+k, cy-r, cx, cy-r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy-r, cx-r, cy-k, cx-r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy+r, cx, cy+r, cx, cy+r))
	p.stream.WriteString("f\n")
}

func (p *pdfBuilder) strokedCircle(cx, cy, r float64) {
	k := r * 0.5523
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f m\n", cx, cy+r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+k, cy+r, cx+r, cy+k, cx+r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+r, cy-k, cx+k, cy-r, cx, cy-r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy-r, cx-r, cy-k, cx-r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy+r, cx, cy+r, cx, cy+r))
	p.stream.WriteString("S\n")
}

func (p *pdfBuilder) filledStrokedCircle(cx, cy, r float64) {
	k := r * 0.5523
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f m\n", cx, cy+r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+k, cy+r, cx+r, cy+k, cx+r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx+r, cy-k, cx+k, cy-r, cx, cy-r))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy-r, cx-r, cy-k, cx-r, cy))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f %.2f %.2f c\n", cx-k, cy+r, cx, cy+r, cx, cy+r))
	p.stream.WriteString("B\n") // fill and stroke
}

// textCentered draws text centered at (cx, cy) using the current font.
func (p *pdfBuilder) textCentered(cx, cy float64, text string, fontSize float64, font string) {
	// Approximate width: 0.6 * fontSize per character for Helvetica
	w := float64(len(text)) * fontSize * 0.6
	x := cx - w/2
	y := cy - fontSize*0.35
	p.stream.WriteString("BT\n")
	p.stream.WriteString(fmt.Sprintf("/%s %.1f Tf\n", font, fontSize))
	p.stream.WriteString(fmt.Sprintf("%.2f %.2f Td\n", x, y))
	p.stream.WriteString(fmt.Sprintf("(%s) Tj\n", pdfEscape(text)))
	p.stream.WriteString("ET\n")
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// drawX draws an X marker at (cx, cy) in PDF coordinates.
func (p *pdfBuilder) drawX(cx, cy float64, size float64) {
	p.setLineWidth(2)
	p.line(cx-size, cy-size, cx+size, cy+size)
	p.line(cx-size, cy+size, cx+size, cy-size)
	p.setLineWidth(1)
}

// drawChordAt draws a single chord diagram at offset (ox, oy) in PDF coordinates.
// PDF coordinates: origin at bottom-left, y increases upward.
// We convert from our top-down layout by flipping y: pdfY = pageHeight - screenY.
func drawChordAt(pb *pdfBuilder, chord Chord, ox, oy int, pageHeight int) {
	numStrings := len(chord.Frets)
	cw := chordWidth(numStrings)

	gridLeft := ox + margin
	gridTop := oy + topMargin + margin // screen coords (top-down)

	// Helper to convert screen Y to PDF Y
	py := func(screenY int) float64 {
		return float64(pageHeight - screenY)
	}
	px := func(screenX int) float64 {
		return float64(screenX)
	}

	// Draw horizontal lines (frets)
	pb.setStrokeColor(0, 0, 0)
	for row := range gridRows {
		y := gridTop + row*cellHeight
		lw := 1.0
		if row == 0 {
			lw = 3.0 // nut is thicker
		}
		pb.setLineWidth(lw)
		pb.line(px(gridLeft), py(y), px(gridLeft+(numStrings-1)*cellWidth), py(y))
	}
	pb.setLineWidth(1)

	// Draw vertical lines (strings)
	for col := range numStrings {
		x := gridLeft + col*cellWidth
		pb.line(px(x), py(gridTop), px(x), py(gridTop+(gridRows-1)*cellHeight))
	}

	// Draw X (muted) and O (open) markers above the nut
	for i, fd := range chord.Frets {
		cx := gridLeft + i*cellWidth
		cy := gridTop - 15

		if fd.Muted {
			pb.setStrokeColor(0, 0, 0)
			pb.drawX(px(cx), py(cy), 6)
		} else if fd.Fret == 0 {
			if fd.Note != "" {
				pb.setFillColor(1, 0, 0)
				pb.textCentered(px(cx), py(cy), fd.Note, 18, "F2")
			} else {
				pb.setStrokeColor(0, 0, 0)
				pb.setLineWidth(1.5)
				pb.strokedCircle(px(cx), py(cy), 6)
			}
		}
	}

	// Draw finger dots on the grid
	for i, fd := range chord.Frets {
		if fd.Muted || fd.Fret == 0 {
			continue
		}
		cx := gridLeft + i*cellWidth
		cy := gridTop + (fd.Fret-1)*cellHeight + cellHeight/2

		if fd.Note != "" {
			pb.setFillColor(1, 0, 0)
			pb.textCentered(px(cx), py(cy), fd.Note, 18, "F2")
		} else {
			pb.setFillColor(0, 0, 0)
			pb.filledCircle(px(cx), py(cy), float64(dotRadius))
		}
	}

	// Draw chord name above the diagram
	pb.setFillColor(0, 0, 0)
	pb.textCentered(float64(ox+cw/2), py(gridTop-35), chord.Name, 16, "F2")

	// Draw start fret number if not at nut
	if chord.StartFret > 0 {
		pb.setFillColor(0, 0, 0)
		pb.textCentered(px(gridLeft-20), py(gridTop), fmt.Sprintf("%d", chord.StartFret), 10, "F1")
	}
}

// writePDF writes a minimal PDF with the given page dimensions and content stream (single page).
func writePDF(filename string, pageWidth, pageHeight int, content string) error {
	return writeMultiPagePDF(filename, pageWidth, pageHeight, []string{content}, "")
}

// writeMultiPagePDF writes a PDF with multiple pages, each with its own content stream.
func writeMultiPagePDF(filename string, pageWidth, pageHeight int, pages []string, sourceData string) error {
	var objects []string
	numPages := len(pages)

	// Object 1: Catalog (Info reference added if sourceData present)
	objects = append(objects, "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages (kids list built below)
	// Object 3: Font (Helvetica)
	// Object 4: Font (Helvetica-Bold)
	// Then pairs of (Page, Content) for each page

	// Build kids references: pages start at object 5, each page is 2 objects (page + stream)
	var kids []string
	for i := range numPages {
		pageObjNum := 5 + i*2
		kids = append(kids, fmt.Sprintf("%d 0 R", pageObjNum))
	}
	objects = append(objects, fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), numPages))

	// Object 3: Font (Helvetica)
	objects = append(objects, "3 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")
	// Object 4: Font (Helvetica-Bold)
	objects = append(objects, "4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n")

	// Page + Content stream pairs
	for i, content := range pages {
		pageObjNum := 5 + i*2
		contentObjNum := 6 + i*2
		objects = append(objects, fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Contents %d 0 R /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> >>\nendobj\n", pageObjNum, pageWidth, pageHeight, contentObjNum))
		objects = append(objects, fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", contentObjNum, len(content), content))
	}

	// Info dictionary with source data as metadata
	if sourceData != "" {
		infoObjNum := len(objects) + 1
		escaped := pdfEscape(sourceData)
		objects = append(objects, fmt.Sprintf("%d 0 obj\n<< /Subject (%s) >>\nendobj\n", infoObjNum, escaped))
		// Patch catalog to reference Info
		objects[0] = fmt.Sprintf("1 0 obj\n<< /Type /Catalog /Pages 2 0 R /Info %d 0 R >>\nendobj\n", infoObjNum)
	}

	var buf strings.Builder
	buf.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = buf.Len()
		buf.WriteString(obj)
	}

	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	buf.WriteString("trailer\n")
	buf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(objects)+1))
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	buf.WriteString("%%EOF\n")

	return os.WriteFile(filename, []byte(buf.String()), 0644)
}

// DrawChordDiagram generates a PDF file for a single chord.
func DrawChordDiagram(chord Chord, filename string, sourceData string) error {
	numStrings := len(chord.Frets)
	if numStrings == 0 {
		return fmt.Errorf("chord has no fret definitions")
	}

	imgWidth := chordWidth(numStrings)
	imgHeight := chordHeight()

	pb := &pdfBuilder{}
	drawChordAt(pb, chord, 0, 0, imgHeight)

	return writeMultiPagePDF(filename, imgWidth, imgHeight, []string{pb.stream.String()}, sourceData)
}

// DrawChordPage generates a multi-page PDF with chord diagrams arranged in
// 4 columns with 1-inch margins, filling columns first then wrapping to rows.
// Subsequent pages are created when rows overflow an A4 page.
func DrawChordPage(chords []Chord, title string, filename string, sourceData string) error {
	if len(chords) == 0 {
		return fmt.Errorf("no chords to draw")
	}

	const (
		pageWidth  = 595 // A4 width in points
		pageHeight = 842 // A4 height in points
		inchMargin = 72  // 1 inch = 72 points
		cols       = 4
	)

	// Available drawing area
	availWidth := pageWidth - 2*inchMargin
	availHeight := pageHeight - 2*inchMargin

	// Native chord dimensions
	cw := chordWidth(6)
	ch := chordHeight()

	// Scale factor so 4 chords fit across the available width
	colWidth := float64(availWidth) / float64(cols)
	scale := colWidth / float64(cw)

	// Scaled chord height determines how many rows fit per page
	scaledHeight := float64(ch) * scale
	rows := int(float64(availHeight) / scaledHeight)
	if rows < 1 {
		rows = 1
	}
	chordsPerPage := cols * rows

	var pages []string

	for start := 0; start < len(chords); start += chordsPerPage {
		end := start + chordsPerPage
		if end > len(chords) {
			end = len(chords)
		}
		pageChords := chords[start:end]

		pb := &pdfBuilder{}

		// Draw title at top of first page
		if start == 0 && title != "" {
			pb.setFillColor(0, 0, 0)
			titleY := float64(pageHeight - inchMargin/2)
			pb.textCentered(float64(pageWidth)/2, titleY, title, 20, "F2")
		}

		for i, chord := range pageChords {
			row := i / cols
			col := i % cols

			// Position in page coordinates (points)
			pdfX := float64(inchMargin) + float64(col)*colWidth
			// The chord in local coords occupies y from 0 to ch (bottom to top).
			// After scaling, the top of the chord is at pdfY + scaledHeight.
			// We want the top of row 0 at pageHeight-inchMargin, so:
			pdfY := float64(pageHeight-inchMargin) - float64(row+1)*scaledHeight

			// Use PDF cm operator to translate and scale
			pb.stream.WriteString("q\n") // save graphics state
			pb.stream.WriteString(fmt.Sprintf("%.4f 0 0 %.4f %.2f %.2f cm\n",
				scale, scale, pdfX, pdfY))

			// Draw chord at origin (0,0) in the scaled coordinate space
			// In this space, the chord draws from (0, -chordHeight) to (chordWidth, 0)
			// since drawChordAt uses pageHeight to flip y, we set pageHeight = ch
			// so that screen y=0 maps to PDF y=ch (top) and screen y=ch maps to PDF y=0 (bottom)
			drawChordAt(pb, chord, 0, 0, ch)

			pb.stream.WriteString("Q\n") // restore graphics state
		}

		pages = append(pages, pb.stream.String())
	}

	return writeMultiPagePDF(filename, pageWidth, pageHeight, pages, sourceData)
}
