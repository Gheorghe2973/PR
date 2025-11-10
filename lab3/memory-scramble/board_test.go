package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Test Strategy:
// - ParseFromFile: valid file, invalid dimensions, wrong card count
// - Flip: all rules 1-A through 3-B
// - Look: perspective for different players
// - Map: consistent transformation of matching cards
// - Concurrency: multiple players, waiting, race conditions

func createTestBoard(t *testing.T, content string) *Board {
	tmpfile, err := os.CreateTemp("", "board*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	board, err := ParseFromFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return board
}

// Test ParseFromFile with valid board
func TestParseFromFile_Valid(t *testing.T) {
	content := "2x2\nA\nB\nA\nB\n"
	board := createTestBoard(t, content)
	
	if board.rows != 2 || board.cols != 2 {
		t.Errorf("Expected 2x2 board, got %dx%d", board.rows, board.cols)
	}
	
	if board.cards[0][0].value != "A" {
		t.Errorf("Expected card at (0,0) to be 'A', got '%s'", board.cards[0][0].value)
	}
}

// Test ParseFromFile with invalid card count
func TestParseFromFile_InvalidCardCount(t *testing.T) {
	content := "2x2\nA\nB\n"
	tmpfile, _ := os.CreateTemp("", "board*.txt")
	tmpfile.Write([]byte(content))
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	
	_, err := ParseFromFile(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid card count")
	}
}

// Rule 1-A: No card at position
func TestFlip_Rule1A_NoCard(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	board.cards[0][0] = nil // Remove card
	
	err := board.Flip("player1", 0, 0)
	if err == nil {
		t.Error("Expected error when flipping empty position")
	}
}

// Rule 1-B: Flip face-down card
func TestFlip_Rule1B_FaceDown(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	err := board.Flip("player1", 0, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	card := board.cards[0][0]
	if card.status != FaceUp {
		t.Error("Card should be face up")
	}
	if card.controller != "player1" {
		t.Errorf("Expected controller 'player1', got '%s'", card.controller)
	}
}

// Rule 1-C: Flip face-up uncontrolled card
func TestFlip_Rule1C_FaceUpUncontrolled(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	// Player1 flips and makes move
	board.Flip("player1", 0, 0)
	board.Flip("player1", 0, 1) // no match
	board.Flip("player1", 1, 0) // new turn, card at (0,0) now uncontrolled
	
	// Player2 flips same card (now face up but uncontrolled)
	err := board.Flip("player2", 0, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	card := board.cards[0][0]
	if card.controller != "player2" {
		t.Errorf("Expected controller 'player2', got '%s'", card.controller)
	}
}

// Rule 2-A: Second card is empty
func TestFlip_Rule2A_SecondCardEmpty(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	board.Flip("player1", 0, 0)
	board.cards[0][1] = nil // Remove second card
	
	err := board.Flip("player1", 0, 1)
	if err == nil {
		t.Error("Expected error when second card is empty")
	}
	
	// First card should be released
	if board.cards[0][0].controller != "" {
		t.Error("First card should be uncontrolled after failed second flip")
	}
}

// Rule 2-B: Second card is controlled
func TestFlip_Rule2B_SecondCardControlled(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	board.Flip("player1", 0, 0)
	board.Flip("player2", 0, 1) // Player2 controls this card
	
	err := board.Flip("player1", 0, 1)
	if err == nil {
		t.Error("Expected error when second card is controlled")
	}
}

// Rule 2-D: Cards match
func TestFlip_Rule2D_Match(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	board.Flip("player1", 0, 0) // A
	board.Flip("player1", 1, 0) // A - match!
	
	player := board.players["player1"]
	if !player.matched {
		t.Error("Expected matched to be true")
	}
	
	if board.cards[0][0].controller != "player1" {
		t.Error("Player should control first matched card")
	}
	if board.cards[1][0].controller != "player1" {
		t.Error("Player should control second matched card")
	}
}

// Rule 2-E: Cards don't match
func TestFlip_Rule2E_NoMatch(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	board.Flip("player1", 0, 0) // A
	board.Flip("player1", 0, 1) // B - no match
	
	player := board.players["player1"]
	if player.matched {
		t.Error("Expected matched to be false")
	}
	
	if board.cards[0][0].controller != "" {
		t.Error("First card should be uncontrolled after no match")
	}
	if board.cards[0][1].controller != "" {
		t.Error("Second card should be uncontrolled after no match")
	}
}

// Rule 3-A: Remove matched cards
func TestFlip_Rule3A_RemoveMatched(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	board.Flip("player1", 0, 0)
	board.Flip("player1", 1, 0) // Match
	board.Flip("player1", 0, 1) // New turn - should remove matched cards
	
	if board.cards[0][0] != nil {
		t.Error("Matched card at (0,0) should be removed")
	}
	if board.cards[1][0] != nil {
		t.Error("Matched card at (1,0) should be removed")
	}
}

// Rule 3-B: Turn non-matching cards face down
func TestFlip_Rule3B_FaceDown(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	board.Flip("player1", 0, 0)
	board.Flip("player1", 0, 1) // No match
	board.Flip("player1", 1, 0) // New turn - should turn previous cards face down
	
	if board.cards[0][0].status != FaceDown {
		t.Error("Non-matching card at (0,0) should be face down")
	}
	if board.cards[0][1].status != FaceDown {
		t.Error("Non-matching card at (0,1) should be face down")
	}
}

// Test scenario from spec: P1 flips, P2 flips same card, P1 continues
func TestFlip_Scenario_CardUsedByOthers(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	// P1 flips C1, C2 (no match)
	board.Flip("player1", 0, 0)
	board.Flip("player1", 0, 1)
	
	// P2 flips C1
	board.Flip("player2", 0, 0)
	
	// P1 flips new first card C3
	board.Flip("player1", 1, 0)
	
	// C1 should remain face up (used by P2)
	if board.cards[0][0].status != FaceUp {
		t.Error("Card at (0,0) should remain face up (used by player2)")
	}
	
	// C2 should be face down (not used by others)
	if board.cards[0][1].status != FaceDown {
		t.Error("Card at (0,1) should be face down")
	}
}

// Test Look returns correct board state
func TestLook(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	board.Flip("player1", 0, 0)
	
	state := board.Look("player1")
	lines := strings.Split(strings.TrimSpace(state), "\n")
	
	if lines[0] != "2x2" {
		t.Errorf("Expected '2x2', got '%s'", lines[0])
	}
	
	if lines[1] != "my A" {
		t.Errorf("Expected 'my A', got '%s'", lines[1])
	}
	
	if lines[2] != "down" {
		t.Errorf("Expected 'down', got '%s'", lines[2])
	}
}

// Test Map transforms cards consistently
func TestMap(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nA\nB\n")
	
	board.Map(func(card string) string {
		if card == "A" {
			return "X"
		}
		return card
	})
	
	if board.cards[0][0].value != "X" {
		t.Errorf("Expected 'X', got '%s'", board.cards[0][0].value)
	}
	if board.cards[1][0].value != "X" {
		t.Errorf("Expected 'X', got '%s'", board.cards[1][0].value)
	}
	if board.cards[0][1].value != "B" {
		t.Errorf("Expected 'B', got '%s'", board.cards[0][1].value)
	}
}

// Test concurrent players
func TestConcurrentPlayers(t *testing.T) {
	board := createTestBoard(t, "3x3\nA\nB\nC\nA\nB\nC\nA\nB\nC\n")
	
	done := make(chan bool, 2)
	
	// Player 1
	go func() {
		board.Flip("player1", 0, 0)
		time.Sleep(10 * time.Millisecond)
		board.Flip("player1", 1, 0)
		done <- true
	}()
	
	// Player 2
	go func() {
		board.Flip("player2", 0, 1)
		time.Sleep(10 * time.Millisecond)
		board.Flip("player2", 1, 1)
		done <- true
	}()
	
	<-done
	<-done
	
	// Board should be in valid state
	board.mu.RLock()
	defer board.mu.RUnlock()
	
	for i := 0; i < board.rows; i++ {
		for j := 0; j < board.cols; j++ {
			if board.cards[i][j] != nil {
				// Check no invalid controller
				if board.cards[i][j].controller != "" {
					if _, exists := board.players[board.cards[i][j].controller]; !exists {
						t.Error("Card has invalid controller")
					}
				}
			}
		}
	}
}

// Test waiting for controlled card
func TestWaitForControlledCard(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	board.Flip("player1", 0, 0) // Player1 controls (0,0)
	
	done := make(chan bool)
	
	// Player2 tries to flip same card - should wait
	go func() {
		board.Flip("player2", 0, 0)
		done <- true
	}()
	
	// Give player2 time to start waiting
	time.Sleep(50 * time.Millisecond)
	
	// Player1 makes another move, releasing the card
	board.Flip("player1", 0, 1)
	
	// Player2 should now get the card
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Player2 should have gotten the card")
	}
}

// Test WaitForChange
func TestWaitForChange(t *testing.T) {
	board := createTestBoard(t, "2x2\nA\nB\nC\nD\n")
	
	changed := make(chan bool)
	
	go func() {
		board.WaitForChange()
		changed <- true
	}()
	
	time.Sleep(50 * time.Millisecond)
	
	// Make a change
	board.Flip("player1", 0, 0)
	
	select {
	case <-changed:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("WaitForChange should have been notified")
	}
}