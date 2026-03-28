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

// RawDeck is a parsed decklist with a name and card entries.
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

		if line == "---" {
			if current != nil {
				decks = append(decks, *current)
				current = nil
			}
			continue
		}

		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if current == nil {
			current = &RawDeck{Name: line}
			continue
		}

		spaceIdx := strings.IndexByte(line, ' ')
		if spaceIdx == -1 {
			return nil, fmt.Errorf("invalid card line: %q", line)
		}
		qty, err := strconv.Atoi(line[:spaceIdx])
		if err != nil {
			return nil, fmt.Errorf("invalid quantity in line %q: %w", line, err)
		}
		name := strings.TrimSpace(line[spaceIdx+1:])
		current.Cards = append(current.Cards, CardEntry{Quantity: qty, Name: name})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if current != nil {
		decks = append(decks, *current)
	}

	return decks, nil
}
