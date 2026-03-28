// Package parser reads MTG decklist text files into structured data.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CardEntry represents a single card and its quantity in a decklist.
type CardEntry struct {
	Quantity int
	Name     string
}

// RawDeck represents is a parsed decklist with a name and card entries.
type RawDeck struct {
	Name  string
	Cards []CardEntry
}

// Parse reads a decklist from r and returns the parsed decks.
// Multiple decks in one input are separated by "---" lines.
// Blank lines and lines starting with "//" are ignored.
func Parse(r io.Reader) ([]RawDeck, error) {
	scanner := bufio.NewScanner(r)
	var decks []RawDeck
	var current *RawDeck

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// check for deck separator
		if line == "---" {
			if current != nil {
				decks = append(decks, *current)
				current = nil
			}
			continue
		}

		// Ignore comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// begin new deck
		if current == nil {
			current = &RawDeck{Name: line}
			continue
		}

		// add cards to deck
		card, err := parseCardLine(line)
		if err != nil {
			return nil, err
		}
		current.Cards = append(current.Cards, card)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// add new deck to list of decks
	if current != nil {
		decks = append(decks, *current)
	}

	return decks, nil
}

// parseCardLine splits a line like "2 Lightning Bolt" into a CardEntry.
func parseCardLine(line string) (CardEntry, error) {
	spaceIdx := strings.IndexByte(line, ' ')
	if spaceIdx == -1 {
		return CardEntry{}, fmt.Errorf("invalid card line: %q", line)
	}

	// parse quantity
	qty, err := strconv.Atoi(line[:spaceIdx])
	if err != nil {
		return CardEntry{}, fmt.Errorf("invalid quantity in line %q: %w", line, err)
	}

	// parse name
	name := strings.TrimSpace(line[spaceIdx+1:])
	return CardEntry{Quantity: qty, Name: name}, nil
}
