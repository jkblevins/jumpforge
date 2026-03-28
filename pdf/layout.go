// layout.go defines the spatial dimensions, spacing, color schemes, and page
// composition for decklist cards. It answers "where and how big" — card size,
// margins, grid arrangement, and color-to-scheme mapping. Rendering logic
// (drawing shapes, text, and images) lives in render.go.
package pdf

import (
	"fmt"
	"strings"

	"github.com/signintech/gopdf"

	"jumpforge/deck"
)

// Card dimensions in PDF points (1 inch = 72 points).
const (
	cardW = 178.6 // 63mm (MTG card width)
	cardH = 249.4 // 88mm (MTG card height)

	pageW = 612 // 8.5 inches (US Letter)
	pageH = 792 // 11 inches (US Letter)
)

// Card structure constants.
const (
	outerBorderW = 2.5  // thick outer frame
	innerBorderW = 0.5  // thin inner frame line
	innerInset   = 5.0  // distance from outer edge to inner frame
	colorBarH    = 16.0 // top color identity bar (holds deck title)
	marginX      = 8.0  // text left margin from inner frame
	marginY      = 10.0 // text top margin from color bar
)

// Typography constants.
const (
	fontTitle    = 10.0
	fontHeader   = 8.0
	fontBody     = 7.5
	lineHeight   = 9.0
	indentX      = 5.0 // extra indent for card entries under headers
	groupSpacing = 2.0 // vertical space between type groups
)

// Mana icon constants.
const (
	iconSize = 12.0 // icon dimensions in PDF points
	iconGap  = 2.0  // space between icons
)

// Grid layout constants for batch mode.
const (
	gridCols = 3   // card columns per page in batch mode
	gridRows = 3   // card rows per page in batch mode
	cardGap  = 8.0 // spacing between cards in the grid
)

// colorScheme defines the border/bar color and background tint for a color identity.
type colorScheme struct {
	border [3]uint8 // color bar and outer border
	bg     [3]uint8 // card background fill
}

// colorMap maps single-letter color identities to their visual scheme.
var colorMap = map[string]colorScheme{
	"W": {border: [3]uint8{170, 145, 80}, bg: [3]uint8{245, 240, 225}},  // White: darker gold border, warm cream bg
	"U": {border: [3]uint8{14, 104, 171}, bg: [3]uint8{215, 232, 245}},  // Blue: blue border, light blue bg
	"B": {border: [3]uint8{50, 40, 50}, bg: [3]uint8{225, 220, 225}},    // Black: near-black border, light gray-purple bg
	"R": {border: [3]uint8{211, 32, 41}, bg: [3]uint8{245, 225, 220}},   // Red: red border, light pink bg
	"G": {border: [3]uint8{0, 115, 62}, bg: [3]uint8{220, 238, 220}},    // Green: green border, light green bg
	"M": {border: [3]uint8{170, 145, 80}, bg: [3]uint8{245, 238, 220}},  // Multicolor: darker gold border, warm bg
	"C": {border: [3]uint8{158, 158, 158}, bg: [3]uint8{235, 235, 235}}, // Colorless: gray border, light gray bg
}

// RenderSingle creates a card-sized PDF containing a single decklist card
// and writes it to outPath.
func RenderSingle(d deck.Deck, outPath string) error {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: cardW, H: cardH},
	})
	if err := setupFonts(pdf); err != nil {
		return fmt.Errorf("setup fonts: %w", err)
	}
	pdf.AddPage()
	renderCard(pdf, d, 0, 0)
	return pdf.WritePdf(outPath)
}

// RenderBatch creates a letter-sized PDF with decklist cards arranged in a
// 3x3 grid. If more than 9 decks are provided, additional pages are added.
func RenderBatch(decks []deck.Deck, outPath string) error {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: pageW, H: pageH},
	})
	if err := setupFonts(pdf); err != nil {
		return fmt.Errorf("setup fonts: %w", err)
	}

	perPage := gridCols * gridRows

	// Center the grid on the page, accounting for gaps between cards.
	gridW := float64(gridCols)*cardW + float64(gridCols-1)*cardGap
	gridH := float64(gridRows)*cardH + float64(gridRows-1)*cardGap
	offsetX := (pageW - gridW) / 2
	offsetY := (pageH - gridH) / 2

	for i, d := range decks {
		if i%perPage == 0 {
			pdf.AddPage()
		}
		slot := i % perPage
		col := slot % gridCols
		row := slot / gridCols
		x := offsetX + float64(col)*(cardW+cardGap)
		y := offsetY + float64(row)*(cardH+cardGap)
		renderCard(pdf, d, x, y)
	}

	return pdf.WritePdf(outPath)
}

// OutputFileName converts a deck name to a PDF filename by lowercasing,
// replacing spaces with dashes, and appending ".pdf".
func OutputFileName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".pdf"
}
