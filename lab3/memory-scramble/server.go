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
	board, err = LoadBoardFromFile("perfect.txt")
	if err != nil {
		log.Fatal("Error loading board:", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/look/", handleLook)
	http.HandleFunc("/flip/", handleFlip)
	http.HandleFunc("/watch/", handleWatch)
	http.HandleFunc("/replace/", handleReplace)

	port := "8080"
	fmt.Printf("Memory Scramble server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleLook(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	playerID := parts[2]

	log.Printf("Look request from player: %s", playerID)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}

func handleFlip(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	playerID := parts[2]
	coords := strings.Split(parts[3], ",")
	if len(coords) != 2 {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	row, err1 := strconv.Atoi(coords[0])
	col, err2 := strconv.Atoi(coords[1])

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid coordinates format", http.StatusBadRequest)
		return
	}

	log.Printf("Flip request from player %s at (%d, %d)", playerID, row, col)

	playerState := board.GetPlayerState(playerID)

	board.mu.Lock()

	if row < 0 || row >= board.Rows || col < 0 || col >= board.Cols {
		board.mu.Unlock()
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	card := &board.Cards[row][col]

	var success bool
	if !playerState.HasFirst {
		log.Printf("Player %s flipping FIRST card at (%d, %d)", playerID, row, col)
		CleanupPreviousPlay(board, playerState, playerID)
		success = FlipFirstCard(board, card, row, col, playerID, playerState)
		if !success {
			board.mu.Unlock()
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Cannot flip that card", http.StatusConflict)
			return
		}
	} else {
		log.Printf("Player %s flipping SECOND card at (%d, %d)", playerID, row, col)
		success = FlipSecondCard(board, card, row, col, playerID, playerState)
		if !success {
			board.mu.Unlock()
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Cannot flip that card", http.StatusConflict)
			return
		}
	}

	board.version++
	board.mu.Unlock()

	board.NotifyListeners()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}

func handleWatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	playerID := parts[2]

	log.Printf("Watch request from player: %s", playerID)

	ch := make(chan struct{}, 1)

	board.listenersMu.Lock()
	board.listeners[ch] = true
	board.listenersMu.Unlock()

	defer func() {
		board.listenersMu.Lock()
		delete(board.listeners, ch)
		board.listenersMu.Unlock()
		close(ch)
	}()

	select {
	case <-ch:
	case <-time.After(30 * time.Second):
	case <-r.Context().Done():
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}

func handleReplace(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	playerID := parts[2]
	fromCard := parts[3]
	toCard := parts[4]

	log.Printf("Replace request from player %s: %s -> %s", playerID, fromCard, toCard)

	board.mu.Lock()
	replaced := ReplaceCards(board, playerID, fromCard, toCard)
	if replaced {
		board.version++
	}
	board.mu.Unlock()

	if replaced {
		board.NotifyListeners()
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, board.FormatBoard(playerID))
}
