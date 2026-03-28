package pdf

import (
	"fmt"
	"strings"

	"github.com/signintech/gopdf"

	"jumpforge/deck"
)

// gridCols is the number of card columns per page in batch mode.
const gridCols = 3

// gridRows is the number of card rows per page in batch mode.
const gridRows = 3

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

	// Center the grid on the page.
	offsetX := (pageW - float64(gridCols)*cardW) / 2
	offsetY := (pageH - float64(gridRows)*cardH) / 2

	for i, d := range decks {
		if i%perPage == 0 {
			pdf.AddPage()
		}
		slot := i % perPage
		col := slot % gridCols
		row := slot / gridCols
		x := offsetX + float64(col)*cardW
		y := offsetY + float64(row)*cardH
		renderCard(pdf, d, x, y)
	}

	return pdf.WritePdf(outPath)
}

// OutputFileName converts a deck name to a PDF filename by lowercasing,
// replacing spaces with dashes, and appending ".pdf".
func OutputFileName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".pdf"
}
