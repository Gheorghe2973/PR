package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var board *Board

func main() {
	var err error

	// Încarcă tabla din fișier
	board, err = LoadBoardFromFile("perfect.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Configurează endpoints
	http.HandleFunc("/look/", handleLook)
	http.HandleFunc("/flip/", handleFlip)
	http.HandleFunc("/watch/", handleWatch)
	http.HandleFunc("/replace/", handleReplace)
	http.Handle("/", http.FileServer(http.Dir(".")))

	// Pornește serverul
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleLook servește request-uri GET /look/{playerID}
// Returnează starea curentă a tablei pentru un jucător
//
// Specification:
//
//	HTTP Method: GET
//	URL Pattern: /look/{playerID}
//	Parameters:
//	  - playerID: identificatorul jucătorului (din URL)
//	Response:
//	  - Status: 200 OK
//	  - Content-Type: text/plain
//	  - Body: output-ul lui board.FormatBoard(playerID)
//	Preconditions:
//	  - board != nil (global)
//	Effects:
//	  - Citește din board (thread-safe)
//	  - Trimite răspuns HTTP
func handleLook(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimPrefix(r.URL.Path, "/look/")

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Fprint(w, board.FormatBoard(playerID))
}

// handleFlip servește request-uri GET /flip/{playerID}/{row},{col}
// Întoarce o carte la poziția specificată
//
// Specification:
//
//	HTTP Method: GET
//	URL Pattern: /flip/{playerID}/{row},{col}
//	Parameters:
//	  - playerID: identificatorul jucătorului (din URL)
//	  - row: rândul cărții (0-indexed, din URL)
//	  - col: coloana cărții (0-indexed, din URL)
//	Response:
//	  - Dacă operația reușește:
//	      - Status: 200 OK
//	      - Content-Type: text/plain
//	      - Body: starea tablei după flip
//	  - Dacă operația eșuează:
//	      - Status: 409 Conflict
//	      - Body: "Cannot flip that card"
//	Preconditions:
//	  - board != nil (global)
//	  - 0 <= row < board.Rows
//	  - 0 <= col < board.Cols
//	Postconditions:
//	  - Dacă reușește:
//	      - board.version este incrementat
//	      - Toți listeners sunt notificați
//	      - Cartea și playerState sunt modificate conform regulilor
//	Effects:
//	  - Modifică board.Cards (thread-safe cu board.mu)
//	  - Incrementează board.version
//	  - Notifică listeners
//	  - Trimite răspuns HTTP
func handleFlip(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /flip/player1/0,1
	path := strings.TrimPrefix(r.URL.Path, "/flip/")
	parts := strings.Split(path, "/")
	playerID := parts[0]
	coords := strings.Split(parts[1], ",")
	row, _ := strconv.Atoi(coords[0])
	col, _ := strconv.Atoi(coords[1])

	playerState := board.GetPlayerState(playerID)

	// Lock pentru scriere (exclusiv)
	board.mu.Lock()
	defer board.mu.Unlock()

	var success bool
	card := &board.Cards[row][col]

	if !playerState.HasFirst {
		// Curăță tura anterioară înainte de a începe una nouă
		CleanupPreviousPlay(board, playerState, playerID)
		// Încearcă să întoarcă prima carte
		success = FlipFirstCard(board, card, row, col, playerID, playerState)
	} else {
		// Încearcă să întoarcă a doua carte
		success = FlipSecondCard(board, card, row, col, playerID, playerState)
	}

	if !success {
		// Operația a eșuat
		http.Error(w, "Cannot flip that card", http.StatusConflict)
		return
	}

	// Incrementează versiunea și notifică listeners
	board.version++
	board.NotifyListeners()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}

// handleWatch servește request-uri GET /watch/{playerID}
// Long polling - blochează până când tabla se modifică
//
// Specification:
//
//	HTTP Method: GET
//	URL Pattern: /watch/{playerID}
//	Parameters:
//	  - playerID: identificatorul jucătorului (din URL)
//	Response:
//	  - Status: 200 OK
//	  - Content-Type: text/plain
//	  - Body: starea tablei după modificare
//	Preconditions:
//	  - board != nil (global)
//	Postconditions:
//	  - Funcția blochează până când:
//	      1. Tabla se modifică (primește notificare), SAU
//	      2. Timeout de 30 secunde, SAU
//	      3. Clientul se deconectează
//	Effects:
//	  - Adaugă canal la board.listeners (thread-safe)
//	  - Șterge canal din board.listeners la final
//	  - Blochează thread-ul curent
//	  - Citește din board
//	  - Trimite răspuns HTTP
func handleWatch(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimPrefix(r.URL.Path, "/watch/")

	// Creează un canal pentru notificări
	ch := make(chan struct{}, 1)

	// Adaugă canalul la listeners
	board.listenersMu.Lock()
	board.listeners[ch] = true
	board.listenersMu.Unlock()

	// La final, șterge canalul din listeners
	defer func() {
		board.listenersMu.Lock()
		delete(board.listeners, ch)
		board.listenersMu.Unlock()
		close(ch)
	}()

	// Așteaptă până când:
	// 1. Tabla se schimbă (primește semnal pe canal)
	// 2. Timeout de 30 secunde
	// 3. Clientul se deconectează
	select {
	case <-ch:
		// Tabla s-a schimbat
	case <-time.After(30 * time.Second):
		// Timeout
	case <-r.Context().Done():
		// Client deconectat
		return
	}

	// Returnează starea actualizată
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}

// handleReplace servește request-uri GET /replace/{playerID}/{from}/{to}
// Înlocuiește toate cărțile controlate de jucător cu o valoare nouă
//
// Specification:
//
//	HTTP Method: GET
//	URL Pattern: /replace/{playerID}/{from}/{to}
//	Parameters:
//	  - playerID: identificatorul jucătorului (din URL)
//	  - from: valoarea de înlocuit (din URL)
//	  - to: noua valoare (din URL)
//	Response:
//	  - Status: 200 OK
//	  - Content-Type: text/plain
//	  - Body: starea tablei după înlocuire
//	Preconditions:
//	  - board != nil (global)
//	Postconditions:
//	  - Dacă cel puțin o carte a fost înlocuită:
//	      - board.version este incrementat
//	      - Toți listeners sunt notificați
//	Effects:
//	  - Poate modifica board.Cards (thread-safe cu board.mu)
//	  - Poate incrementa board.version
//	  - Poate notifica listeners
//	  - Trimite răspuns HTTP
func handleReplace(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /replace/player1/A/B
	path := strings.TrimPrefix(r.URL.Path, "/replace/")
	parts := strings.Split(path, "/")
	playerID := parts[0]
	fromCard := parts[1]
	toCard := parts[2]

	// Lock pentru scriere
	board.mu.Lock()
	replaced := ReplaceCards(board, playerID, fromCard, toCard)
	if replaced {
		board.version++
	}
	board.mu.Unlock()

	// Notifică listeners dacă s-au făcut schimbări
	if replaced {
		board.NotifyListeners()
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}
