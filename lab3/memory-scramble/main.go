package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var board *Board

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run . <port> <board-file>")
		os.Exit(1)
	}

	port := os.Args[1]
	boardFile := os.Args[2]

	// Parse board from file
	var err error
	board, err = ParseFromFile(boardFile)
	if err != nil {
		log.Fatalf("Failed to parse board file: %v", err)
	}

	fmt.Printf("Loaded board:\n%s\n", board)
	fmt.Printf("Starting Memory Scramble server on port %s...\n", port)
	fmt.Printf("Open http://localhost:%s in your browser to play!\n", port)

	// Set up HTTP handlers
	http.HandleFunc("/look/", corsMiddleware(handleLook))
	http.HandleFunc("/flip/", corsMiddleware(handleFlip))
	http.HandleFunc("/replace/", corsMiddleware(handleReplace))
	http.HandleFunc("/watch/", corsMiddleware(handleWatch))
	
	// Serve static files
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleLook handles GET /look/<playerID>
func handleLook(w http.ResponseWriter, r *http.Request) {
	playerID := extractPlayerID(r.URL.Path, "/look/")
	if playerID == "" {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	boardState := board.Look(playerID)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, boardState)
}

// handleFlip handles GET /flip/<playerID>/<row>,<col>
func handleFlip(w http.ResponseWriter, r *http.Request) {
	// Extract player ID and coordinates
	path := strings.TrimPrefix(r.URL.Path, "/flip/")
	parts := strings.Split(path, "/")
	
	if len(parts) != 2 {
		http.Error(w, "Invalid flip request", http.StatusBadRequest)
		return
	}

	playerID := parts[0]
	coords := parts[1]

	// Parse coordinates
	coordParts := strings.Split(coords, ",")
	if len(coordParts) != 2 {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	row, err := strconv.Atoi(coordParts[0])
	if err != nil {
		http.Error(w, "Invalid row", http.StatusBadRequest)
		return
	}

	col, err := strconv.Atoi(coordParts[1])
	if err != nil {
		http.Error(w, "Invalid column", http.StatusBadRequest)
		return
	}

	// Perform flip
	err = board.Flip(playerID, row, col)
	if err != nil {
		// Return 409 Conflict for flip failures per spec
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, err.Error())
		return
	}

	// Return updated board state
	boardState := board.Look(playerID)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, boardState)
}

// handleReplace handles GET /replace/<playerID>/<fromCard>/<toCard>
func handleReplace(w http.ResponseWriter, r *http.Request) {
	// Extract parameters
	path := strings.TrimPrefix(r.URL.Path, "/replace/")
	parts := strings.SplitN(path, "/", 3)
	
	if len(parts) != 3 {
		http.Error(w, "Invalid replace request", http.StatusBadRequest)
		return
	}

	playerID := parts[0]
	fromCard := parts[1]
	toCard := parts[2]

	// Perform map operation
	err := board.Map(func(card string) string {
		if card == fromCard {
			return toCard
		}
		return card
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated board state
	boardState := board.Look(playerID)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, boardState)
}

// handleWatch handles GET /watch/<playerID>
func handleWatch(w http.ResponseWriter, r *http.Request) {
	playerID := extractPlayerID(r.URL.Path, "/watch/")
	if playerID == "" {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	// Wait for board change
	board.WaitForChange()

	// Return updated board state
	boardState := board.Look(playerID)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, boardState)
}

// extractPlayerID extracts player ID from URL path
func extractPlayerID(path, prefix string) string {
	playerID := strings.TrimPrefix(path, prefix)
	playerID = strings.Split(playerID, "/")[0]
	
	// Validate player ID: alphanumeric and underscore only
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, playerID); !matched {
		return ""
	}
	
	return playerID
}