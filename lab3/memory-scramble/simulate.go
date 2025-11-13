package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	players := []string{"player1", "player2", "player3", "player4"}
	movesPerPlayer := 100

	fmt.Println("Starting simulation with 4 players, 100 moves each...")
	fmt.Println("Timeouts: 0.1ms - 2ms")

	var wg sync.WaitGroup

	for _, player := range players {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			simulatePlayer(p, movesPerPlayer)
		}(player)
	}

	wg.Wait()
	fmt.Println("\nSimulation complete - no crashes!")
}

func simulatePlayer(playerID string, moves int) {
	for i := 0; i < moves; i++ {
		// Random timeout between 0.1ms and 2ms
		timeout := time.Duration(rand.Float64()*1.9+0.1) * time.Millisecond
		time.Sleep(timeout)

		// Random position
		row := rand.Intn(3)
		col := rand.Intn(3)

		// Make flip request
		url := fmt.Sprintf("http://localhost:8080/flip/%s/%d,%d", playerID, row, col)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("[%s] Error on move %d: %v\n", playerID, i+1, err)
			continue
		}

		// Read and discard response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Print progress every 25 moves
		if (i+1)%25 == 0 {
			fmt.Printf("[%s] Completed %d/%d moves\n", playerID, i+1, moves)
		}
	}
	fmt.Printf("[%s] Finished all %d moves\n", playerID, moves)
}
