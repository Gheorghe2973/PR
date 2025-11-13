package main

// Card represents a single card in the Memory Scramble game.
type Card struct {
	Value      string
	FaceUp     bool
	Controller string
}

func NewCard(value string) Card {
	return Card{
		Value:      value,
		FaceUp:     false,
		Controller: "",
	}
}

func (c *Card) checkRep() {
	if c.Value == "" {
		if c.FaceUp {
			panic("Eliminated card cannot be face-up")
		}
		if c.Controller != "" {
			panic("Eliminated card cannot be controlled")
		}
	}
	if c.Controller != "" && !c.FaceUp {
		panic("Controlled card must be face-up")
	}
}
