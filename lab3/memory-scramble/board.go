package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Board represents a Memory Scramble game board.
type Board struct {
	Rows           int
	Cols           int
	Cards          [][]Card
	mu             sync.RWMutex
	version        int
	listeners      map[chan struct{}]bool
	listenersMu    sync.Mutex
	playerStates   map[string]*PlayerState
	playerStatesMu sync.Mutex
}

// LoadBoardFromFile loads a game board from a file.
func LoadBoardFromFile(filename string) (*Board, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}

	dimensions := scanner.Text()
	var rows, cols int
	_, err = fmt.Sscanf(dimensions, "%dx%d", &rows, &cols)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}

	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("dimensions must be positive")
	}

	cards := make([][]Card, rows)
	for i := 0; i < rows; i++ {
		cards[i] = make([]Card, cols)
		for j := 0; j < cols; j++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("not enough cards")
			}
			value := strings.TrimSpace(scanner.Text())
			cards[i][j] = NewCard(value)
		}
	}

	board := &Board{
		Rows:         rows,
		Cols:         cols,
		Cards:        cards,
		version:      0,
		listeners:    make(map[chan struct{}]bool),
		playerStates: make(map[string]*PlayerState),
	}

	board.checkRep()
	return board, nil
}

func (b *Board) checkRep() {
	if b.Rows <= 0 || b.Cols <= 0 {
		panic("Board dimensions must be positive")
	}
	if len(b.Cards) != b.Rows {
		panic(fmt.Sprintf("Cards length %d doesn't match Rows %d", len(b.Cards), b.Rows))
	}
	for i := 0; i < b.Rows; i++ {
		if len(b.Cards[i]) != b.Cols {
			panic(fmt.Sprintf("Row %d length %d doesn't match Cols %d", i, len(b.Cards[i]), b.Cols))
		}
		for j := 0; j < b.Cols; j++ {
			b.Cards[i][j].checkRep()
		}
	}
	if b.version < 0 {
		panic("Version cannot be negative")
	}
}

func (b *Board) GetPlayerState(playerID string) *PlayerState {
	b.playerStatesMu.Lock()
	defer b.playerStatesMu.Unlock()

	if _, exists := b.playerStates[playerID]; !exists {
		b.playerStates[playerID] = NewPlayerState()
	}
	return b.playerStates[playerID]
}

func (b *Board) NotifyListeners() {
	b.listenersMu.Lock()
	defer b.listenersMu.Unlock()

	for ch := range b.listeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (b *Board) FormatBoard(playerID string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%dx%d\n", b.Rows, b.Cols))

	for i := 0; i < b.Rows; i++ {
		for j := 0; j < b.Cols; j++ {
			card := b.Cards[i][j]

			if card.Value == "" {
				result.WriteString("none\n")
			} else if !card.FaceUp {
				result.WriteString("down\n")
			} else if card.Controller == playerID {
				result.WriteString(fmt.Sprintf("my %s\n", card.Value))
			} else {
				result.WriteString(fmt.Sprintf("up %s\n", card.Value))
			}
		}
	}

	return result.String()
}
