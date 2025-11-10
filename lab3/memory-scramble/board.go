package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

/*
 * Board - Memory Scramble Game Board ADT
 * 
 * Abstraction Function:
 *   AF(rows, cols, cards, players) = game board of size rows x cols
 *     where cards[i][j] is the card at position (i,j) or nil if empty
 * 
 * Representation Invariant:
 *   - rows > 0 && cols > 0
 *   - len(cards) == rows && for all i: len(cards[i]) == cols
 *   - for all players: if firstCard != nil then it's a valid position
 * 
 * Safety from Rep Exposure:
 *   - All fields are private
 *   - cards array never returned directly
 *   - Mutex protects concurrent access
 */

type CardStatus int

const (
	FaceDown CardStatus = iota
	FaceUp
)

type Card struct {
	value         string
	status        CardStatus
	controller    string
	lastTouchedBy string
}

type PlayerState struct {
	id          string
	firstCard   *Position
	secondCard  *Position
	matched     bool
}

type Position struct {
	row int
	col int
}

type Board struct {
	rows     int
	cols     int
	cards    [][]*Card
	mu       sync.RWMutex
	players  map[string]*PlayerState
	waiters  map[Position][]chan bool
	watchers []chan bool
}

// ParseFromFile creates a board from a file
// Format: ROWSxCOLS followed by ROWS*COLS card values
func ParseFromFile(filename string) (*Board, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}
	
	dims := strings.Split(scanner.Text(), "x")
	if len(dims) != 2 {
		return nil, fmt.Errorf("invalid dimensions format")
	}
	
	rows, err := strconv.Atoi(dims[0])
	if err != nil {
		return nil, fmt.Errorf("invalid rows: %v", err)
	}
	
	cols, err := strconv.Atoi(dims[1])
	if err != nil {
		return nil, fmt.Errorf("invalid columns: %v", err)
	}
	
	cards := make([][]*Card, rows)
	for i := 0; i < rows; i++ {
		cards[i] = make([]*Card, cols)
	}
	
	cardIndex := 0
	for scanner.Scan() {
		cardValue := strings.TrimSpace(scanner.Text())
		if cardValue == "" {
			continue
		}
		
		row := cardIndex / cols
		col := cardIndex % cols
		
		if row >= rows {
			return nil, fmt.Errorf("too many cards in file")
		}
		
		cards[row][col] = &Card{
			value:         cardValue,
			status:        FaceDown,
			controller:    "",
			lastTouchedBy: "",
		}
		
		cardIndex++
	}
	
	if cardIndex != rows*cols {
		return nil, fmt.Errorf("expected %d cards, got %d", rows*cols, cardIndex)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	
	return &Board{
		rows:     rows,
		cols:     cols,
		cards:    cards,
		players:  make(map[string]*PlayerState),
		waiters:  make(map[Position][]chan bool),
		watchers: make([]chan bool, 0),
	}, nil
}

func (b *Board) getOrCreatePlayer(playerID string) *PlayerState {
	if player, exists := b.players[playerID]; exists {
		return player
	}
	
	player := &PlayerState{
		id:         playerID,
		firstCard:  nil,
		secondCard: nil,
		matched:    false,
	}
	b.players[playerID] = player
	return player
}

func (b *Board) isValidPosition(pos Position) bool {
	return pos.row >= 0 && pos.row < b.rows && pos.col >= 0 && pos.col < b.cols
}

func (b *Board) notifyWatchers() {
	for _, ch := range b.watchers {
		select {
		case ch <- true:
		default:
		}
	}
}

func (b *Board) notifyWaiters(pos Position) {
	if waiters, exists := b.waiters[pos]; exists {
		for _, ch := range waiters {
			select {
			case ch <- true:
			default:
			}
		}
		delete(b.waiters, pos)
	}
}

// Flip attempts to flip a card at (row, col) for the given player
// Implements game rules 1-A through 3-B
func (b *Board) Flip(playerID string, row, col int) error {
	b.mu.Lock()
	
	pos := Position{row: row, col: col}
	
	if !b.isValidPosition(pos) {
		b.mu.Unlock()
		return fmt.Errorf("invalid position: (%d, %d)", row, col)
	}
	
	player := b.getOrCreatePlayer(playerID)
	
	// Rules 3-A and 3-B: cleanup from previous turn
	if player.firstCard != nil && player.secondCard != nil {
		if !player.matched {
			card1 := b.cards[player.firstCard.row][player.firstCard.col]
			if card1 != nil && card1.controller == playerID {
				card1.controller = ""
			}
			
			card2 := b.cards[player.secondCard.row][player.secondCard.col]
			if card2 != nil && card2.controller == playerID {
				card2.controller = ""
			}
		}
		
		if player.matched {
			// Rule 3-A: remove matched cards
			b.cards[player.firstCard.row][player.firstCard.col] = nil
			b.cards[player.secondCard.row][player.secondCard.col] = nil
			b.notifyWatchers()
		} else {
			// Rule 3-B: turn face down if not used by others
			card1 := b.cards[player.firstCard.row][player.firstCard.col]
			if card1 != nil && card1.status == FaceUp && card1.controller == "" && card1.lastTouchedBy == playerID {
				card1.status = FaceDown
				b.notifyWatchers()
			}
			
			card2 := b.cards[player.secondCard.row][player.secondCard.col]
			if card2 != nil && card2.status == FaceUp && card2.controller == "" && card2.lastTouchedBy == playerID {
				card2.status = FaceDown
				b.notifyWatchers()
			}
		}
		
		player.firstCard = nil
		player.secondCard = nil
		player.matched = false
	}
	
	card := b.cards[row][col]
	
	// Flipping first card
	if player.firstCard == nil {
		// Rule 1-A: no card there
		if card == nil {
			b.mu.Unlock()
			return fmt.Errorf("no card at position (%d, %d)", row, col)
		}
		
		// Rule 1-D: wait for controlled card
		if card.controller != "" && card.controller != playerID {
			waitChan := make(chan bool, 1)
			b.waiters[pos] = append(b.waiters[pos], waitChan)
			b.mu.Unlock()
			
			<-waitChan
			
			b.mu.Lock()
			card = b.cards[row][col]
			if card == nil {
				b.mu.Unlock()
				return fmt.Errorf("no card at position (%d, %d)", row, col)
			}
		}
		
		// Rule 1-B/1-C: turn face up and take control
		if card.status == FaceDown {
			card.status = FaceUp
			b.notifyWatchers()
		}
		card.controller = playerID
		card.lastTouchedBy = playerID
		player.firstCard = &pos
		
		b.mu.Unlock()
		return nil
	}
	
	// Flipping second card
	
	// Rule 2-A: no card there
	if card == nil {
		card1 := b.cards[player.firstCard.row][player.firstCard.col]
		if card1 != nil {
			card1.controller = ""
			b.notifyWaiters(*player.firstCard)
		}
		player.firstCard = nil
		b.mu.Unlock()
		return fmt.Errorf("no card at position (%d, %d)", row, col)
	}
	
	// Rule 2-B: card is controlled
	if card.controller != "" {
		card1 := b.cards[player.firstCard.row][player.firstCard.col]
		if card1 != nil {
			card1.controller = ""
			b.notifyWaiters(*player.firstCard)
		}
		player.firstCard = nil
		b.mu.Unlock()
		return fmt.Errorf("card at position (%d, %d) is controlled", row, col)
	}
	
	// Rule 2-C: turn face up if needed
	wasChanged := false
	if card.status == FaceDown {
		card.status = FaceUp
		wasChanged = true
	}
	
	player.secondCard = &pos
	card.lastTouchedBy = playerID
	card1 := b.cards[player.firstCard.row][player.firstCard.col]
	
	// Rule 2-D/2-E: check match
	if card1.value == card.value {
		card.controller = playerID
		player.matched = true
	} else {
		card1.controller = ""
		card.controller = ""
		b.notifyWaiters(*player.firstCard)
		player.matched = false
	}
	
	if wasChanged {
		b.notifyWatchers()
	}
	
	b.mu.Unlock()
	return nil
}

// Look returns the board state from player's perspective
// Format: ROWSxCOLS\n(SPOT\n)+ where SPOT is "none", "down", "up VALUE", or "my VALUE"
func (b *Board) Look(playerID string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	_ = b.getOrCreatePlayer(playerID)
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("%dx%d\n", b.rows, b.cols))
	
	for i := 0; i < b.rows; i++ {
		for j := 0; j < b.cols; j++ {
			card := b.cards[i][j]
			
			if card == nil {
				result.WriteString("none\n")
			} else if card.status == FaceDown {
				result.WriteString("down\n")
			} else if card.controller == playerID {
				result.WriteString(fmt.Sprintf("my %s\n", card.value))
			} else {
				result.WriteString(fmt.Sprintf("up %s\n", card.value))
			}
		}
	}
	
	return result.String()
}

// Map applies transformer to all cards consistently
func (b *Board) Map(transformer func(string) string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	valueMap := make(map[string]string)
	
	for i := 0; i < b.rows; i++ {
		for j := 0; j < b.cols; j++ {
			card := b.cards[i][j]
			if card != nil {
				if _, exists := valueMap[card.value]; !exists {
					valueMap[card.value] = transformer(card.value)
				}
			}
		}
	}
	
	changed := false
	for i := 0; i < b.rows; i++ {
		for j := 0; j < b.cols; j++ {
			card := b.cards[i][j]
			if card != nil {
				newValue := valueMap[card.value]
				if newValue != card.value {
					card.value = newValue
					changed = true
				}
			}
		}
	}
	
	if changed {
		b.notifyWatchers()
	}
	
	return nil
}

// WaitForChange blocks until board changes
func (b *Board) WaitForChange() {
	b.mu.Lock()
	ch := make(chan bool, 1)
	b.watchers = append(b.watchers, ch)
	b.mu.Unlock()
	
	<-ch
	
	b.mu.Lock()
	for i, watcher := range b.watchers {
		if watcher == ch {
			b.watchers = append(b.watchers[:i], b.watchers[i+1:]...)
			break
		}
	}
	b.mu.Unlock()
}

// String returns human-readable board representation
func (b *Board) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Board %dx%d:\n", b.rows, b.cols))
	
	for i := 0; i < b.rows; i++ {
		for j := 0; j < b.cols; j++ {
			card := b.cards[i][j]
			if card == nil {
				result.WriteString("[   ] ")
			} else if card.status == FaceDown {
				result.WriteString("[ ? ] ")
			} else {
				result.WriteString(fmt.Sprintf("[%3s] ", card.value))
			}
		}
		result.WriteString("\n")
	}
	
	return result.String()
}