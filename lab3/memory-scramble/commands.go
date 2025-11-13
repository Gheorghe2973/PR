package main

import (
	"log"
)

// FlipFirstCard implements rules 1-A, 1-B, 1-C, 1-D
func FlipFirstCard(board *Board, card *Card, row, col int, playerID string, playerState *PlayerState) bool {
	// Rule 1-A: No card there
	if card.Value == "" {
		log.Printf("Rule 1-A: No card at (%d, %d)", row, col)
		return false
	}

	// Rule 1-B: Card is face-down
	if !card.FaceUp {
		log.Printf("Rule 1-B: Turning up card at (%d, %d)", row, col)
		card.FaceUp = true
		card.Controller = playerID
		playerState.HasFirst = true
		playerState.FirstCardRow = row
		playerState.FirstCardCol = col
		card.checkRep()
		playerState.checkRep()
		return true
	}

	// Rule 1-C: Card is face-up but not controlled
	if card.Controller == "" {
		log.Printf("Rule 1-C: Taking control of card at (%d, %d)", row, col)
		card.Controller = playerID
		playerState.HasFirst = true
		playerState.FirstCardRow = row
		playerState.FirstCardCol = col
		card.checkRep()
		playerState.checkRep()
		return true
	}

	// Rule 1-D: Card is controlled by another player
	log.Printf("Rule 1-D: Card at (%d, %d) is controlled by %s", row, col, card.Controller)
	return false
}

// FlipSecondCard implements rules 2-A, 2-B, 2-C, 2-D, 2-E
func FlipSecondCard(board *Board, card *Card, row, col int, playerID string, playerState *PlayerState) bool {
	// Rule 2-A: No card there
	if card.Value == "" {
		log.Printf("Rule 2-A: No card at (%d, %d), relinquishing first card", row, col)
		firstCard := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
		firstCard.Controller = ""
		playerState.HasFirst = false
		playerState.HasSecond = false
		firstCard.checkRep()
		playerState.checkRep()
		return false
	}

	// Rule 2-B: Card is controlled
	if card.FaceUp && card.Controller != "" {
		log.Printf("Rule 2-B: Card at (%d, %d) is controlled, relinquishing first card", row, col)
		firstCard := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
		firstCard.Controller = ""
		playerState.HasFirst = false
		playerState.HasSecond = false
		firstCard.checkRep()
		playerState.checkRep()
		return false
	}

	// Rule 2-C: Turn face-up if needed
	if !card.FaceUp {
		log.Printf("Rule 2-C: Turning up card at (%d, %d)", row, col)
		card.FaceUp = true
	}

	firstCard := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]

	// Rule 2-D: Cards match
	if firstCard.Value == card.Value {
		log.Printf("Rule 2-D: Match! %s == %s", firstCard.Value, card.Value)
		card.Controller = playerID
		playerState.HasSecond = true
		playerState.SecondCardRow = row
		playerState.SecondCardCol = col
		playerState.Matched = true
		playerState.HasFirst = false
		card.checkRep()
		playerState.checkRep()
		return true
	}

	// Rule 2-E: No match
	log.Printf("Rule 2-E: No match. %s != %s", firstCard.Value, card.Value)
	firstCard.Controller = ""
	card.Controller = ""
	playerState.HasSecond = true
	playerState.SecondCardRow = row
	playerState.SecondCardCol = col
	playerState.Matched = false
	playerState.HasFirst = false
	firstCard.checkRep()
	card.checkRep()
	playerState.checkRep()
	return true
}

// CleanupPreviousPlay implements rule 3-A and 3-B
func CleanupPreviousPlay(board *Board, playerState *PlayerState, playerID string) {
	log.Printf("Cleaning up previous play for player %s (HasFirst: %v, HasSecond: %v, Matched: %v)",
		playerID, playerState.HasFirst, playerState.HasSecond, playerState.Matched)

	if playerState.HasSecond {
		if playerState.Matched {
			// Rule 3-A: Remove matched cards
			card1 := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
			card2 := &board.Cards[playerState.SecondCardRow][playerState.SecondCardCol]

			if card1.Controller == playerID {
				log.Printf("Removing matched card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
				card1.Value = ""
				card1.FaceUp = false
				card1.Controller = ""
				card1.checkRep()
			}
			if card2.Controller == playerID {
				log.Printf("Removing matched card at (%d, %d)", playerState.SecondCardRow, playerState.SecondCardCol)
				card2.Value = ""
				card2.FaceUp = false
				card2.Controller = ""
				card2.checkRep()
			}
		} else {
			// Rule 3-B: Turn down non-matched cards
			card1 := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
			card2 := &board.Cards[playerState.SecondCardRow][playerState.SecondCardCol]

			if card1.Value != "" && card1.FaceUp && card1.Controller == "" {
				log.Printf("Turning down card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
				card1.FaceUp = false
				card1.checkRep()
			}
			if card2.Value != "" && card2.FaceUp && card2.Controller == "" {
				log.Printf("Turning down card at (%d, %d)", playerState.SecondCardRow, playerState.SecondCardCol)
				card2.FaceUp = false
				card2.checkRep()
			}
		}
	} else if playerState.HasFirst {
		// Rule 3-B: Turn down single card
		card1 := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
		if card1.Value != "" && card1.FaceUp && card1.Controller == "" {
			log.Printf("Turning down single card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
			card1.FaceUp = false
			card1.checkRep()
		}
	}

	playerState.HasFirst = false
	playerState.HasSecond = false
	playerState.Matched = false
	playerState.checkRep()
}

// ReplaceCards replaces all cards controlled by player with matching value
func ReplaceCards(board *Board, playerID, fromCard, toCard string) bool {
	replaced := false
	for i := 0; i < board.Rows; i++ {
		for j := 0; j < board.Cols; j++ {
			card := &board.Cards[i][j]
			if card.Controller == playerID && card.Value == fromCard && card.FaceUp {
				card.Value = toCard
				replaced = true
				card.checkRep()
				log.Printf("Replaced card at (%d, %d)", i, j)
			}
		}
	}
	return replaced
}
