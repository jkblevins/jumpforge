package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"jumpforge/deck"
)

func testDeck(name, color string) deck.Deck {
	return deck.Deck{
		Name:          name,
		DominantColor: color,
		Groups: []deck.TypeGroup{
			{
				TypeName: "Creature",
				Count:    3,
				Cards: []deck.DeckCard{
					{Name: "Llanowar Elves", Quantity: 2, CMC: 1},
					{Name: "Tarmogoyf", Quantity: 1, CMC: 2},
				},
			},
			{
				TypeName: "Sorcery",
				Count:    2,
				Cards: []deck.DeckCard{
					{Name: "Rampant Growth", Quantity: 2, CMC: 2},
				},
			},
			{
				TypeName: "Land",
				Count:    10,
				Cards: []deck.DeckCard{
					{Name: "Forest", Quantity: 7, CMC: 0},
					{Name: "Stomping Ground", Quantity: 3, CMC: 0},
				},
			},
		},
	}
}

func TestRenderSingleCard(t *testing.T) {
	d := testDeck("Gruul Smash", "G")
	out := filepath.Join(t.TempDir(), "single.pdf")

	if err := RenderSingle(d, out); err != nil {
		t.Fatalf("RenderSingle: %v", err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output PDF is empty")
	}
}

func TestRenderBatchGrid(t *testing.T) {
	decks := []deck.Deck{
		testDeck("Gruul Smash", "G"),
		testDeck("Dimir Control", "U"),
	}
	out := filepath.Join(t.TempDir(), "batch.pdf")

	if err := RenderBatch(decks, out); err != nil {
		t.Fatalf("RenderBatch: %v", err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output PDF is empty")
	}
}
