// render.go implements the drawing logic for decklist cards. It answers
// "how to draw" — shapes, text, images, and colors onto the PDF surface.
// Spatial dimensions and color schemes are defined in layout.go.
package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"

	"github.com/signintech/gopdf"

	"jumpforge/deck"
)

//go:embed fonts/DejaVuSans.ttf
var fontRegular []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var fontBold []byte

//go:embed icons/W.png
var iconW []byte

//go:embed icons/U.png
var iconU []byte

//go:embed icons/B.png
var iconB []byte

//go:embed icons/R.png
var iconR []byte

//go:embed icons/G.png
var iconG []byte

//go:embed icons/C.png
var iconC []byte

// manaIcons maps color identity letters to their embedded PNG data.
var manaIcons = map[string][]byte{
	"W": iconW,
	"U": iconU,
	"B": iconB,
	"R": iconR,
	"G": iconG,
	"C": iconC,
}

// setupFonts registers the embedded DejaVu Sans regular and bold fonts with the PDF.
func setupFonts(pdf *gopdf.GoPdf) error {
	if err := pdf.AddTTFFontData("body", fontRegular); err != nil {
		return fmt.Errorf("add regular font: %w", err)
	}
	if err := pdf.AddTTFFontDataWithOption("body", fontBold, gopdf.TtfOption{Style: gopdf.Bold}); err != nil {
		return fmt.Errorf("add bold font: %w", err)
	}
	return nil
}

// pluralType converts a singular card type name to its plural form for
// display in group headers.
func pluralType(t string) string {
	switch t {
	case "Sorcery":
		return "Sorceries"
	case "Land":
		return "Lands"
	default:
		return t + "s"
	}
}

// cardLine formats a card entry: singles as "Card Name", multiples as "Card Name (N)".
func cardLine(c deck.DeckCard) string {
	if c.Quantity == 1 {
		return c.Name
	}
	return fmt.Sprintf("%s (%d)", c.Name, c.Quantity)
}

// decodeManaIcon decodes an embedded PNG into an image.Image.
func decodeManaIcon(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}

// cardRenderer holds the state needed to draw a single decklist card.
type cardRenderer struct {
	pdf    *gopdf.GoPdf
	x, y   float64    // card origin (upper-left corner)
	curY   float64     // vertical cursor for text placement
	scheme colorScheme
}

// drawFrame draws the outer border, tinted background, and inner frame line.
func (cr *cardRenderer) drawFrame() {
	b := cr.scheme.border
	bg := cr.scheme.bg

	// Tinted background with color-matched border.
	// Inset by half the stroke width so the border stays within card bounds.
	half := outerBorderW / 2
	cr.pdf.SetFillColor(bg[0], bg[1], bg[2])
	cr.pdf.SetStrokeColor(b[0], b[1], b[2])
	cr.pdf.SetLineWidth(outerBorderW)
	cr.pdf.Rectangle(cr.x+half, cr.y+half, cr.x+cardW-half, cr.y+cardH-half, "DF", 0, 0)

	// Inner frame line in a darker shade of the border color.
	cr.pdf.SetStrokeColor(b[0]/2, b[1]/2, b[2]/2)
	cr.pdf.SetLineWidth(innerBorderW)
	cr.pdf.Rectangle(cr.x+innerInset, cr.y+innerInset, cr.x+cardW-innerInset, cr.y+cardH-innerInset, "D", 0, 0)
}

// drawColorBar fills the top bar with the deck's border color, inside the inner frame.
func (cr *cardRenderer) drawColorBar() {
	b := cr.scheme.border
	cr.pdf.SetFillColor(b[0], b[1], b[2])
	barX := cr.x + innerInset + innerBorderW
	barY := cr.y + innerInset + innerBorderW
	barW := cardW - 2*(innerInset+innerBorderW)
	cr.pdf.RectFromUpperLeftWithStyle(barX, barY, barW, colorBarH, "F")
}

// drawTitle renders the deck name left-aligned in white on top of the color bar.
func (cr *cardRenderer) drawTitle(name string) {
	cr.pdf.SetFont("body", "B", fontTitle)
	cr.pdf.SetTextColor(255, 255, 255)

	barY := cr.y + innerInset + innerBorderW
	nameX := cr.x + innerInset + marginX
	nameY := barY + (colorBarH+fontTitle*0.65)/2
	cr.pdf.SetXY(nameX, nameY)
	cr.pdf.Text(name)

	// Reset text color for subsequent content.
	cr.pdf.SetTextColor(30, 30, 30)
}

// drawColorIdentity renders mana symbol icons right-aligned in the color bar.
func (cr *cardRenderer) drawColorIdentity(colors []string) {
	barY := cr.y + innerInset + innerBorderW
	iconY := barY + (colorBarH-iconSize)/2

	// Calculate total width to right-align.
	totalW := float64(len(colors))*iconSize + float64(len(colors)-1)*iconGap
	startX := cr.x + cardW - innerInset - marginX/2 - totalW

	for i, color := range colors {
		data, ok := manaIcons[color]
		if !ok {
			continue
		}
		img, err := decodeManaIcon(data)
		if err != nil {
			continue
		}
		x := startX + float64(i)*(iconSize+iconGap)
		cr.pdf.ImageFrom(img, x, iconY, &gopdf.Rect{W: iconSize, H: iconSize})
	}
}

// drawGroups renders all type groups with headers and indented card entries.
func (cr *cardRenderer) drawGroups(groups []deck.TypeGroup) {
	textLeft := cr.x + innerInset + marginX

	for _, g := range groups {
		header := fmt.Sprintf("%s:", pluralType(g.TypeName))
		cr.pdf.SetFont("body", "B", fontHeader)
		cr.pdf.SetTextColor(30, 30, 30)
		cr.pdf.SetXY(textLeft, cr.curY)
		cr.pdf.Text(header)
		cr.curY += lineHeight

		cr.pdf.SetFont("body", "", fontBody)
		cr.pdf.SetTextColor(50, 50, 50)
		for _, c := range g.Cards {
			line := cardLine(c)
			cr.pdf.SetXY(textLeft+indentX, cr.curY)
			cr.pdf.Text(line)
			cr.curY += lineHeight
		}
		cr.curY += groupSpacing
	}
}

// renderCard draws a single decklist card onto the PDF at the given offset.
func renderCard(p *gopdf.GoPdf, d deck.Deck, x, y float64) {
	scheme := colorMap["C"]
	if s, ok := colorMap[d.DominantColor]; ok {
		scheme = s
	}
	cr := &cardRenderer{
		pdf:    p,
		x:      x,
		y:      y,
		curY:   y + innerInset + innerBorderW + colorBarH + marginY,
		scheme: scheme,
	}
	cr.drawFrame()
	cr.drawColorBar()
	cr.drawTitle(d.Name)
	cr.drawColorIdentity(d.ColorIdentity)
	cr.drawGroups(d.Groups)
}
