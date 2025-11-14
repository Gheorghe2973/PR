package main

import "log"

// FlipFirstCard încearcă să întoarcă prima carte pentru un jucător
// Implementează regulile 1-A, 1-B, 1-C, 1-D din specificația jocului
//
// Specification:
//
//	Parameters:
//	  - board: pointer către Board (nu trebuie nil)
//	  - card: pointer către Card de întors (nu trebuie nil)
//	  - row, col: coordonatele cărții în matricea board.Cards
//	  - playerID: identificatorul jucătorului
//	  - playerState: pointer către PlayerState (nu trebuie nil)
//	Returns:
//	  - bool: true dacă operația reușește, false altfel
//	Preconditions:
//	  - board != nil
//	  - card != nil
//	  - playerState != nil
//	  - 0 <= row < board.Rows
//	  - 0 <= col < board.Cols
//	  - card == &board.Cards[row][col]
//	  - playerState.HasFirst == false
//	Postconditions:
//	  - Dacă returnează true:
//	      - card.FaceUp == true
//	      - card.Controller == playerID
//	      - playerState.HasFirst == true
//	      - playerState.FirstCardRow == row
//	      - playerState.FirstCardCol == col
//	  - Dacă returnează false, starea nu se modifică
//	Effects:
//	  - Poate modifica card (FaceUp, Controller)
//	  - Poate modifica playerState (HasFirst, FirstCardRow, FirstCardCol)
//	  - Scrie în log
//	Reguli implementate:
//	  - 1-A: Spațiu gol (Value == "") → false
//	  - 1-B: Carte cu fața în jos → true, întoarce cartea
//	  - 1-C: Carte vizibilă necontrolată → true, preia control
//	  - 1-D: Carte controlată de altcineva → false
func FlipFirstCard(board *Board, card *Card, row, col int, playerID string, playerState *PlayerState) bool {

	// Regula 1-A: Nu există carte la această poziție
	if card.Value == "" {
		log.Printf("Rule 1-A: No card at (%d, %d)", row, col)
		return false
	}

	// Regula 1-B: Carte cu fața în jos - o întoarcem
	if !card.FaceUp {
		log.Printf("Rule 1-B: Turning up card at (%d, %d)", row, col)
		card.FaceUp = true
		card.Controller = playerID
		playerState.HasFirst = true
		playerState.FirstCardRow = row
		playerState.FirstCardCol = col
		return true
	}

	// Regula 1-C: Carte vizibilă dar necontrolată - preluăm controlul
	if card.Controller == "" {
		log.Printf("Rule 1-C: Taking control of card at (%d, %d)", row, col)
		card.Controller = playerID
		playerState.HasFirst = true
		playerState.FirstCardRow = row
		playerState.FirstCardCol = col
		return true
	}

	// Regula 1-D: Carte controlată de altcineva - nu putem face nimic
	log.Printf("Rule 1-D: Card at (%d, %d) is controlled by %s", row, col, card.Controller)
	return false
}

// FlipSecondCard încearcă să întoarcă a doua carte pentru un jucător
// Implementează regulile 2-A, 2-B, 2-C, 2-D, 2-E din specificația jocului
//
// Specification:
//
//	Parameters:
//	  - board: pointer către Board (nu trebuie nil)
//	  - card: pointer către Card de întors (nu trebuie nil)
//	  - row, col: coordonatele cărții în matricea board.Cards
//	  - playerID: identificatorul jucătorului
//	  - playerState: pointer către PlayerState (nu trebuie nil)
//	Returns:
//	  - bool: true dacă operația reușește, false altfel
//	Preconditions:
//	  - board != nil
//	  - card != nil
//	  - playerState != nil
//	  - 0 <= row < board.Rows
//	  - 0 <= col < board.Cols
//	  - card == &board.Cards[row][col]
//	  - playerState.HasFirst == true
//	  - Prima carte există la (playerState.FirstCardRow, playerState.FirstCardCol)
//	Postconditions:
//	  - Dacă returnează true:
//	      - card.FaceUp == true
//	      - playerState.HasSecond == true
//	      - playerState.HasFirst == false
//	      - playerState.SecondCardRow == row
//	      - playerState.SecondCardCol == col
//	      - Dacă cărțile se potrivesc:
//	          - card.Controller == playerID
//	          - firstCard.Controller == playerID (neschimbat)
//	          - playerState.Matched == true
//	      - Dacă cărțile NU se potrivesc:
//	          - card.Controller == ""
//	          - firstCard.Controller == ""
//	          - playerState.Matched == false
//	  - Dacă returnează false:
//	      - firstCard.Controller == ""
//	      - playerState.HasFirst == false
//	Effects:
//	  - Poate modifica card (FaceUp, Controller)
//	  - Poate modifica firstCard (Controller)
//	  - Modifică întotdeauna playerState
//	  - Scrie în log
//	Reguli implementate:
//	  - 2-A: Spațiu gol → false, renunță la prima carte
//	  - 2-B: Carte controlată → false, renunță la prima carte
//	  - 2-C: Carte cu fața în jos → o întoarce
//	  - 2-D: Cărți identice → true, MATCH
//	  - 2-E: Cărți diferite → true, NO MATCH
func FlipSecondCard(board *Board, card *Card, row, col int, playerID string, playerState *PlayerState) bool {

	// Obține prima carte
	firstCard := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]

	// Regula 2-A: Nu există carte - renunță la prima carte
	if card.Value == "" {
		log.Printf("Rule 2-A: No card at (%d, %d), relinquishing first card", row, col)
		firstCard.Controller = ""
		playerState.HasFirst = false
		return false
	}

	// Regula 2-B: Carte controlată - renunță la prima carte
	if card.FaceUp && card.Controller != "" {
		log.Printf("Rule 2-B: Card at (%d, %d) is controlled, relinquishing first card", row, col)
		firstCard.Controller = ""
		playerState.HasFirst = false
		return false
	}

	// Regula 2-C: Dacă cartea e cu fața în jos, o întoarcem
	if !card.FaceUp {
		log.Printf("Rule 2-C: Turning up card at (%d, %d)", row, col)
		card.FaceUp = true
	}

	// Salvăm poziția celei de-a doua cărți
	playerState.SecondCardRow = row
	playerState.SecondCardCol = col
	playerState.HasSecond = true
	playerState.HasFirst = false

	// Verificăm dacă cărțile se potrivesc
	if firstCard.Value == card.Value {
		// Regula 2-D: MATCH - ambele cărți rămân controlate
		log.Printf("Rule 2-D: Match! %s == %s", firstCard.Value, card.Value)
		card.Controller = playerID
		playerState.Matched = true
	} else {
		// Regula 2-E: NO MATCH - ambele cărți devin necontrolate dar vizibile
		log.Printf("Rule 2-E: No match. %s != %s", firstCard.Value, card.Value)
		firstCard.Controller = ""
		card.Controller = ""
		playerState.Matched = false
	}

	return true
}

// CleanupPreviousPlay curăță tabla după tura anterioară a jucătorului
// Implementează regulile 3-A și 3-B din specificația jocului
//
// Specification:
//
//	Parameters:
//	  - board: pointer către Board (nu trebuie nil)
//	  - playerState: pointer către PlayerState (nu trebuie nil)
//	  - playerID: identificatorul jucătorului
//	Preconditions:
//	  - board != nil
//	  - playerState != nil
//	  - Dacă playerState.HasFirst == true:
//	      - Prima carte există la (FirstCardRow, FirstCardCol)
//	  - Dacă playerState.HasSecond == true:
//	      - Prima și a doua carte există la pozițiile respective
//	Postconditions:
//	  - playerState.HasFirst == false
//	  - playerState.HasSecond == false
//	  - playerState.Matched == false
//	  - Dacă tura anterioară a avut match (Matched == true):
//	      - Cărțile potrivite controlate de playerID sunt eliminate
//	        (Value="", FaceUp=false, Controller="")
//	  - Dacă tura anterioară NU a avut match (Matched == false):
//	      - Cărțile necontrolate (Controller=="") sunt întoarse cu fața în jos
//	Effects:
//	  - Poate modifica cărțile din board.Cards
//	  - Modifică întotdeauna playerState
//	  - Scrie în log
//	Reguli implementate:
//	  - 3-A: Cărți potrivite → elimină (doar dacă Controller == playerID)
//	  - 3-B: Cărți nepotrivite → întoarce cu fața în jos (doar dacă Controller == "")
func CleanupPreviousPlay(board *Board, playerState *PlayerState, playerID string) {

	log.Printf("Cleaning up previous play for player %s (HasFirst: %v, HasSecond: %v, Matched: %v)",
		playerID, playerState.HasFirst, playerState.HasSecond, playerState.Matched)

	// Dacă jucătorul avea doar prima carte întorsă
	if playerState.HasFirst {
		card1 := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]

		// Regula 3-B: Întoarce cartea cu fața în jos dacă nu e controlată
		if card1.Value != "" && card1.FaceUp && card1.Controller == "" {
			log.Printf("Turning down card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
			card1.FaceUp = false
		}
	}

	// Dacă jucătorul avea două cărți întoarse
	if playerState.HasSecond {
		card1 := &board.Cards[playerState.FirstCardRow][playerState.FirstCardCol]
		card2 := &board.Cards[playerState.SecondCardRow][playerState.SecondCardCol]

		if playerState.Matched {
			// Regula 3-A: Elimină cărțile potrivite
			if card1.Controller == playerID {
				log.Printf("Removing matched card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
				card1.Value = ""
				card1.FaceUp = false
				card1.Controller = ""
			}
			if card2.Controller == playerID {
				log.Printf("Removing matched card at (%d, %d)", playerState.SecondCardRow, playerState.SecondCardCol)
				card2.Value = ""
				card2.FaceUp = false
				card2.Controller = ""
			}
		} else {
			// Regula 3-B: Întoarce cărțile nepotrivite cu fața în jos
			if card1.Value != "" && card1.FaceUp && card1.Controller == "" {
				log.Printf("Turning down card at (%d, %d)", playerState.FirstCardRow, playerState.FirstCardCol)
				card1.FaceUp = false
			}
			if card2.Value != "" && card2.FaceUp && card2.Controller == "" {
				log.Printf("Turning down card at (%d, %d)", playerState.SecondCardRow, playerState.SecondCardCol)
				card2.FaceUp = false
			}
		}
	}

	// Resetează starea jucătorului
	playerState.HasFirst = false
	playerState.HasSecond = false
	playerState.Matched = false
}

// ReplaceCards înlocuiește toate cărțile controlate de jucător cu o valoare nouă
//
// Specification:
//
//	Parameters:
//	  - board: pointer către Board (nu trebuie nil)
//	  - playerID: identificatorul jucătorului
//	  - fromCard: valoarea de înlocuit
//	  - toCard: noua valoare
//	Returns:
//	  - bool: true dacă cel puțin o carte a fost înlocuită, false altfel
//	Preconditions:
//	  - board != nil
//	  - fromCard și toCard pot fi orice string-uri (inclusiv "")
//	Postconditions:
//	  - Toate cărțile cu card.Controller == playerID și card.Value == fromCard
//	    au acum card.Value == toCard
//	  - Returnează true dacă cel puțin o carte a fost modificată
//	Effects:
//	  - Poate modifica Value-ul cărților din board.Cards
func ReplaceCards(board *Board, playerID, fromCard, toCard string) bool {
	replaced := false

	// Parcurge toate cărțile
	for i := 0; i < board.Rows; i++ {
		for j := 0; j < board.Cols; j++ {
			card := &board.Cards[i][j]

			// Dacă cartea e controlată de acest jucător și are valoarea căutată
			if card.Controller == playerID && card.Value == fromCard {
				card.Value = toCard // Înlocuiește valoarea
				replaced = true
			}
		}
	}

	return replaced
}
