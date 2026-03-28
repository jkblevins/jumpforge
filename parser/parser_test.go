package parser

import (
	"strings"
	"testing"
)

func TestParseSingleDeck(t *testing.T) {
	input := `
Goblin Rush

1 Goblin Guide
2 Lightning Bolt
// comment line
3 Mountain
`
	decks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(decks))
	}
	d := decks[0]
	if d.Name != "Goblin Rush" {
		t.Errorf("expected name 'Goblin Rush', got %q", d.Name)
	}
	if len(d.Cards) != 3 {
		t.Errorf("expected 3 card entries, got %d", len(d.Cards))
	}
	if d.Cards[0].Quantity != 1 || d.Cards[0].Name != "Goblin Guide" {
		t.Errorf("unexpected first card: %+v", d.Cards[0])
	}
	if d.Cards[2].Quantity != 3 || d.Cards[2].Name != "Mountain" {
		t.Errorf("unexpected third card: %+v", d.Cards[2])
	}
}

func TestParseBatchDecks(t *testing.T) {
	input := `Goblin Rush

1 Goblin Guide
3 Mountain
---
Forest Friends

1 Llanowar Elves
4 Forest
`
	decks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
	if decks[0].Name != "Goblin Rush" {
		t.Errorf("expected first deck 'Goblin Rush', got %q", decks[0].Name)
	}
	if decks[1].Name != "Forest Friends" {
		t.Errorf("expected second deck 'Forest Friends', got %q", decks[1].Name)
	}
	if len(decks[1].Cards) != 2 {
		t.Errorf("expected 2 cards in second deck, got %d", len(decks[1].Cards))
	}
}

func TestParseSkipsBlanksAndComments(t *testing.T) {
	input := `My Deck

// this is a comment

1 Lightning Bolt

// another comment
2 Mountain
`
	decks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks[0].Cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(decks[0].Cards))
	}
}

func TestParseCardLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CardEntry
		wantErr bool
	}{
		{"normal card", "2 Lightning Bolt", CardEntry{2, "Lightning Bolt"}, false},
		{"single quantity", "1 Goblin Guide", CardEntry{1, "Goblin Guide"}, false},
		{"high quantity", "10 Forest", CardEntry{10, "Forest"}, false},
		{"no space", "Lightning", CardEntry{}, true},
		{"non-numeric quantity", "X Mountain", CardEntry{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCardLine(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseCardLine(%q) = %+v, want %+v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	decks, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decks) != 0 {
		t.Errorf("expected 0 decks, got %d", len(decks))
	}
}
