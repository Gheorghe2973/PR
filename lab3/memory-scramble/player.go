package main

// PlayerState represents the current state of a player in the game.
type PlayerState struct {
	FirstCardRow  int
	FirstCardCol  int
	SecondCardRow int
	SecondCardCol int
	HasFirst      bool
	HasSecond     bool
	Matched       bool
}

func NewPlayerState() *PlayerState {
	return &PlayerState{
		HasFirst:  false,
		HasSecond: false,
		Matched:   false,
	}
}

func (p *PlayerState) checkRep() {
	if p.HasSecond && p.HasFirst {
		panic("Player cannot have both HasFirst and HasSecond true")
	}
}
