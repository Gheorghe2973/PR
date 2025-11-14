package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Board reprezintă tabla de joc pentru Memory Scramble
// Representation Invariants:
//   - Rows > 0 și Cols > 0
//   - len(Cards) == Rows
//   - Pentru tot i: len(Cards[i]) == Cols
//   - version >= 0
//   - Toate cărțile din Cards respectă Card.checkRep()
//
// Thread Safety:
//   - Cards și version sunt protejate de mu (RWMutex)
//   - listeners este protejat de listenersMu
//   - playerStates este protejat de playerStatesMu
type Board struct {
	Rows           int                     // Numărul de rânduri
	Cols           int                     // Numărul de coloane
	Cards          [][]Card                // Matricea de cărți
	version        int                     // Versiunea tablei (incrementată la fiecare modificare)
	mu             sync.RWMutex            // Protejează Cards și version
	listeners      map[chan struct{}]bool  // Canale pentru watch requests
	listenersMu    sync.Mutex              // Protejează listeners
	playerStates   map[string]*PlayerState // Stările jucătorilor
	playerStatesMu sync.Mutex              // Protejează playerStates
}

// LoadBoardFromFile încarcă tabla de joc dintr-un fișier
//
// Specification:
//
//	Parameters:
//	  - filename: calea către fișierul care conține configurația tablei
//	Returns:
//	  - *Board: pointer către tabla încărcată, sau nil dacă apare eroare
//	  - error: nil dacă operația reușește, altfel eroarea întâlnită
//	Preconditions:
//	  - filename trebuie să existe și să fie citibil
//	  - Fișierul trebuie să aibă formatul:
//	      Linia 1: "RxC" (dimensiuni)
//	      Liniile următoare: R*C valori de cărți
//	Postconditions:
//	  - Dacă reușește: returnează Board valid care respectă invarianții
//	  - Dacă eșuează: returnează nil și error non-nil
//	  - Toate cărțile sunt inițializate cu FaceUp=false, Controller=""
//	Effects:
//	  - Citește din fișierul specificat
//	  - Alocă memorie pentru Board și Cards
func LoadBoardFromFile(filename string) (*Board, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Citește prima linie cu dimensiunile (ex: "3x3")
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}
	dims := strings.Split(scanner.Text(), "x")
	rows, _ := strconv.Atoi(dims[0])
	cols, _ := strconv.Atoi(dims[1])

	// Creează matricea de cărți
	cards := make([][]Card, rows)
	for i := 0; i < rows; i++ {
		cards[i] = make([]Card, cols)
		for j := 0; j < cols; j++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("not enough cards in file")
			}
			// Fiecare carte începe cu fața în jos și necontrolată
			cards[i][j] = Card{
				Value:      scanner.Text(),
				FaceUp:     false,
				Controller: "",
			}
		}
	}

	// Creează Board-ul
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

// checkRep verifică representation invariants pentru Board
//
// Specification:
//
//	Preconditions: none
//	Postconditions:
//	  - Dacă invarianții sunt violați, funcția face panic
//	  - Dacă invarianții sunt respectați, funcția returnează normal
//	Effects:
//	  - Poate face panic dacă invarianții sunt violați
//	  - Verifică și invarianții tuturor cărților din Cards
func (b *Board) checkRep() {
	// Verifică că dimensiunile sunt pozitive
	if b.Rows <= 0 || b.Cols <= 0 {
		panic("Board dimensions must be positive")
	}
	// Verifică că matricea are numărul corect de rânduri
	if len(b.Cards) != b.Rows {
		panic("Cards length doesn't match Rows")
	}
	// Verifică că fiecare rând are numărul corect de coloane
	for i := 0; i < b.Rows; i++ {
		if len(b.Cards[i]) != b.Cols {
			panic("Row length doesn't match Cols")
		}
		// Verifică invarianții fiecărei cărți
		for j := 0; j < b.Cols; j++ {
			b.Cards[i][j].checkRep()
		}
	}
	// Verifică că version nu este negativ
	if b.version < 0 {
		panic("Version cannot be negative")
	}
}

// GetPlayerState returnează sau creează starea pentru un jucător
//
// Specification:
//
//	Parameters:
//	  - playerID: identificatorul unic al jucătorului
//	Returns:
//	  - *PlayerState: pointer către starea jucătorului
//	Preconditions:
//	  - playerID nu trebuie să fie string gol
//	Postconditions:
//	  - Returnează pointer către PlayerState pentru playerID
//	  - Dacă playerID nu exista, creează o stare nouă
//	  - Obiectul returnat respectă PlayerState invariants
//	Thread Safety:
//	  - Funcția este thread-safe (folosește playerStatesMu)
//	Effects:
//	  - Poate modifica playerStates (adaugă entry nou)
//	  - Folosește lock pe playerStatesMu
func (b *Board) GetPlayerState(playerID string) *PlayerState {
	b.playerStatesMu.Lock()
	defer b.playerStatesMu.Unlock()

	// Dacă jucătorul nu există, creează-i o stare nouă
	if _, exists := b.playerStates[playerID]; !exists {
		b.playerStates[playerID] = NewPlayerState()
	}
	return b.playerStates[playerID]
}

// FormatBoard formatează tabla pentru afișare către un jucător specific
//
// Specification:
//
//	Parameters:
//	  - playerID: identificatorul jucătorului pentru care se formatează
//	Returns:
//	  - string: reprezentare text a tablei în formatul:
//	      Linia 1: "RxC"
//	      Linii următoare: pentru fiecare carte:
//	        "none" - carte eliminată
//	        "down" - carte cu fața în jos
//	        "my X" - carte controlată de acest jucător cu valoarea X
//	        "up X" - carte vizibilă cu valoarea X (altcineva sau nimeni)
//	Preconditions:
//	  - playerID poate fi orice string (inclusiv gol)
//	Postconditions:
//	  - Returnează string format corect
//	  - Nu modifică starea Board-ului
//	Thread Safety:
//	  - Funcția este thread-safe (folosește mu.RLock)
//	Effects:
//	  - Citește din Cards (read lock pe mu)
func (b *Board) FormatBoard(playerID string) string {
	b.mu.RLock() // Lock pentru citire (permite citiri simultane)
	defer b.mu.RUnlock()

	var result strings.Builder

	// Prima linie: dimensiunile
	result.WriteString(fmt.Sprintf("%dx%d\n", b.Rows, b.Cols))

	// Pentru fiecare carte, scrie starea ei
	for i := 0; i < b.Rows; i++ {
		for j := 0; j < b.Cols; j++ {
			card := b.Cards[i][j]

			if card.Value == "" {
				// Carte eliminată
				result.WriteString("none\n")
			} else if !card.FaceUp {
				// Carte cu fața în jos
				result.WriteString("down\n")
			} else if card.Controller == playerID {
				// Carte controlată de acest jucător
				result.WriteString(fmt.Sprintf("my %s\n", card.Value))
			} else {
				// Carte vizibilă (controlată de altcineva sau necontrolată)
				result.WriteString(fmt.Sprintf("up %s\n", card.Value))
			}
		}
	}

	return result.String()
}

// NotifyListeners notifică toți listeners că tabla s-a modificat
//
// Specification:
//
//	Preconditions: none
//	Postconditions:
//	  - Trimite semnal pe toate canalele din listeners
//	  - Nu blochează dacă un canal este plin (folosește select cu default)
//	Thread Safety:
//	  - Funcția este thread-safe (folosește listenersMu)
//	Effects:
//	  - Trimite mesaje pe canale
//	  - Folosește lock pe listenersMu
func (b *Board) NotifyListeners() {
	b.listenersMu.Lock()
	defer b.listenersMu.Unlock()

	// Trimite semnal pe fiecare canal
	for ch := range b.listeners {
		select {
		case ch <- struct{}{}: // Încearcă să trimită
			// Success
		default:
			// Canalul e plin, skip
		}
	}
}
