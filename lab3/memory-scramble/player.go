package main

// PlayerState ține evidența stării unui jucător în timpul jocului
// Representation Invariants:
//   - Dacă HasSecond == true atunci HasFirst == false
type PlayerState struct {
	FirstCardRow  int  // Rândul primei cărți (-1 dacă nu există)
	FirstCardCol  int  // Coloana primei cărți (-1 dacă nu există)
	SecondCardRow int  // Rândul celei de-a doua cărți (-1 dacă nu există)
	SecondCardCol int  // Coloana celei de-a doua cărți (-1 dacă nu există)
	HasFirst      bool // true dacă jucătorul are prima carte întorsă
	HasSecond     bool // true dacă jucătorul are a doua carte întorsă
	Matched       bool // true dacă cele două cărți se potrivesc
}

// NewPlayerState creează o stare nouă pentru un jucător
//
// Specification:
//
//	Preconditions: none
//	Postconditions:
//	  - Returnează un pointer către PlayerState nou alocat
//	  - Toate câmpurile row/col sunt setate la -1
//	  - Toate câmpurile bool sunt setate la false
//	  - Obiectul returnat respectă representation invariants
func NewPlayerState() *PlayerState {
	return &PlayerState{
		FirstCardRow:  -1,
		FirstCardCol:  -1,
		SecondCardRow: -1,
		SecondCardCol: -1,
		HasFirst:      false,
		HasSecond:     false,
		Matched:       false,
	}
}

// checkRep verifică representation invariants pentru PlayerState
//
// Specification:
//
//	Preconditions: none
//	Postconditions:
//	  - Dacă invarianții sunt violați, funcția face panic
//	  - Dacă invarianții sunt respectați, funcția returnează normal
//	Effects:
//	  - Poate face panic dacă HasSecond == true și HasFirst == true
func (p *PlayerState) checkRep() {
	if p.HasSecond && p.HasFirst {
		panic("Player cannot have both HasFirst and HasSecond true")
	}
}
