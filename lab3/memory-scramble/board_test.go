package main

import (
	"testing"
)

// Test Rule 1-A: Flipping empty space fails
func TestFlipEmptySpace(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{NewCard("A"), Card{Value: "", FaceUp: false, Controller: ""}},
			{NewCard("B"), NewCard("C")},
		},
	}

	playerState := NewPlayerState()
	card := &board.Cards[0][1]

	success := FlipFirstCard(board, card, 0, 1, "player1", playerState)

	if success {
		t.Error("Expected flip to fail on empty space")
	}
	if playerState.HasFirst {
		t.Error("Player should not have first card after failed flip")
	}
}

// Test Rule 1-B: Flipping face-down card
func TestFlipFaceDownCard(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{NewCard("A"), NewCard("B")},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := NewPlayerState()
	card := &board.Cards[0][0]

	success := FlipFirstCard(board, card, 0, 0, "player1", playerState)

	if !success {
		t.Error("Expected flip to succeed on face-down card")
	}
	if !card.FaceUp {
		t.Error("Card should be face-up after flip")
	}
	if card.Controller != "player1" {
		t.Errorf("Card should be controlled by player1, got %s", card.Controller)
	}
	if !playerState.HasFirst {
		t.Error("Player should have first card")
	}
}

// Test Rule 1-C: Taking control of face-up uncontrolled card
func TestTakeControlOfFaceUpCard(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: ""}, NewCard("B")},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := NewPlayerState()
	card := &board.Cards[0][0]

	success := FlipFirstCard(board, card, 0, 0, "player1", playerState)

	if !success {
		t.Error("Expected flip to succeed on face-up uncontrolled card")
	}
	if card.Controller != "player1" {
		t.Errorf("Card should be controlled by player1, got %s", card.Controller)
	}
	if !playerState.HasFirst {
		t.Error("Player should have first card")
	}
}

// Test Rule 1-D: Cannot flip card controlled by another player
func TestCannotFlipControlledCard(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player2"}, NewCard("B")},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := NewPlayerState()
	card := &board.Cards[0][0]

	success := FlipFirstCard(board, card, 0, 0, "player1", playerState)

	if success {
		t.Error("Expected flip to fail on card controlled by another player")
	}
	if playerState.HasFirst {
		t.Error("Player should not have first card after failed flip")
	}
}

// Test Rule 2-D: Matching cards
func TestMatchingCards(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player1"}, NewCard("A")},
			{NewCard("B"), NewCard("C")},
		},
	}

	playerState := &PlayerState{
		HasFirst:     true,
		FirstCardRow: 0,
		FirstCardCol: 0,
	}
	card := &board.Cards[0][1]

	success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

	if !success {
		t.Error("Expected flip to succeed")
	}
	if !playerState.Matched {
		t.Error("Cards should be matched")
	}
	if !playerState.HasSecond {
		t.Error("Player should have second card")
	}
	if card.Controller != "player1" {
		t.Error("Matched card should be controlled by player")
	}
}

// Test Rule 2-E: Non-matching cards
func TestNonMatchingCards(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player1"}, NewCard("B")},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := &PlayerState{
		HasFirst:     true,
		FirstCardRow: 0,
		FirstCardCol: 0,
	}
	card := &board.Cards[0][1]

	success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

	if !success {
		t.Error("Expected flip to succeed")
	}
	if playerState.Matched {
		t.Error("Cards should not be matched")
	}
	if !playerState.HasSecond {
		t.Error("Player should have second card")
	}

	firstCard := &board.Cards[0][0]
	if firstCard.Controller != "" {
		t.Error("First card should not be controlled after non-match")
	}
	if card.Controller != "" {
		t.Error("Second card should not be controlled after non-match")
	}
}

// Test Rule 3-A: Removing matched cards
func TestRemoveMatchedCards(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player1"}, Card{Value: "A", FaceUp: true, Controller: "player1"}},
			{NewCard("B"), NewCard("C")},
		},
	}

	playerState := &PlayerState{
		HasFirst:      false,
		HasSecond:     true,
		Matched:       true,
		FirstCardRow:  0,
		FirstCardCol:  0,
		SecondCardRow: 0,
		SecondCardCol: 1,
	}

	CleanupPreviousPlay(board, playerState, "player1")

	card1 := &board.Cards[0][0]
	card2 := &board.Cards[0][1]

	if card1.Value != "" {
		t.Error("First matched card should be removed")
	}
	if card2.Value != "" {
		t.Error("Second matched card should be removed")
	}
	if playerState.HasSecond {
		t.Error("Player state should be reset")
	}
}

// Test Rule 3-B: Turning down non-matched cards
func TestTurnDownNonMatched(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: ""}, Card{Value: "B", FaceUp: true, Controller: ""}},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := &PlayerState{
		HasFirst:      false,
		HasSecond:     true,
		Matched:       false,
		FirstCardRow:  0,
		FirstCardCol:  0,
		SecondCardRow: 0,
		SecondCardCol: 1,
	}

	CleanupPreviousPlay(board, playerState, "player1")

	card1 := &board.Cards[0][0]
	card2 := &board.Cards[0][1]

	if card1.FaceUp {
		t.Error("First non-matched card should be face-down")
	}
	if card2.FaceUp {
		t.Error("Second non-matched card should be face-down")
	}
	if playerState.HasSecond {
		t.Error("Player state should be reset")
	}
}

// Test Rule 2-A: No card at position
func TestFlipSecondCardEmptySpace(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player1"}, Card{Value: "", FaceUp: false, Controller: ""}},
			{NewCard("B"), NewCard("C")},
		},
	}

	playerState := &PlayerState{
		HasFirst:     true,
		FirstCardRow: 0,
		FirstCardCol: 0,
	}
	card := &board.Cards[0][1]

	success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

	if success {
		t.Error("Expected flip to fail on empty space")
	}
	if playerState.HasFirst {
		t.Error("Player should relinquish first card after failed second flip")
	}

	firstCard := &board.Cards[0][0]
	if firstCard.Controller != "" {
		t.Error("First card should be relinquished")
	}
}

// Test Rule 2-B: Card already controlled
func TestFlipSecondCardControlled(t *testing.T) {
	board := &Board{
		Rows: 2,
		Cols: 2,
		Cards: [][]Card{
			{Card{Value: "A", FaceUp: true, Controller: "player1"}, Card{Value: "B", FaceUp: true, Controller: "player2"}},
			{NewCard("C"), NewCard("D")},
		},
	}

	playerState := &PlayerState{
		HasFirst:     true,
		FirstCardRow: 0,
		FirstCardCol: 0,
	}
	card := &board.Cards[0][1]

	success := FlipSecondCard(board, card, 0, 1, "player1", playerState)

	if success {
		t.Error("Expected flip to fail on controlled card")
	}
	if playerState.HasFirst {
		t.Error("Player should relinquish first card after failed second flip")
	}
}

// Test loading board from file
func TestLoadBoardFromFile(t *testing.T) {
	board, err := LoadBoardFromFile("perfect.txt")
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	if board.Rows <= 0 || board.Cols <= 0 {
		t.Error("Board dimensions should be positive")
	}

	if len(board.Cards) != board.Rows {
		t.Errorf("Expected %d rows, got %d", board.Rows, len(board.Cards))
	}

	for i, row := range board.Cards {
		if len(row) != board.Cols {
			t.Errorf("Row %d: expected %d cols, got %d", i, board.Cols, len(row))
		}
		for j, card := range row {
			if card.FaceUp {
				t.Errorf("Card at (%d,%d) should start face-down", i, j)
			}
			if card.Controller != "" {
				t.Errorf("Card at (%d,%d) should start uncontrolled", i, j)
			}
		}
	}
}
