package main

// Card reprezintă o carte din jocul Memory Scramble
// Representation Invariants:
//   - Dacă Value == "" atunci FaceUp == false și Controller == ""
//   - Dacă Controller != "" atunci FaceUp == true
type Card struct {
	Value      string // Valoarea cărții sau "" dacă e eliminată
	FaceUp     bool   // true = vizibilă, false = ascunsă
	Controller string // ID-ul jucătorului care o controlează sau ""
}

// checkRep verifică representation invariants pentru Card
//
// Specification:
//
//	Preconditions: none
//	Postconditions:
//	  - Dacă invarianții sunt violați, funcția face panic
//	  - Dacă invarianții sunt respectați, funcția returnează normal
//	Effects:
//	  - Poate face panic dacă invarianții sunt violați
func (c *Card) checkRep() {
	// Verifică că o carte eliminată nu poate fi vizibilă sau controlată
	if c.Value == "" {
		if c.FaceUp || c.Controller != "" {
			panic("Eliminated card cannot be face-up or controlled")
		}
	}
	// Verifică că o carte controlată trebuie să fie vizibilă
	if c.Controller != "" && !c.FaceUp {
		panic("Controlled card must be face-up")
	}
}
