package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Simulation script to test multiple concurrent players
// making random moves with random delays

func simulatePlayer(board *Board, playerID string, moves int, results chan string) {
	defer func() {
		if r := recover(); r != nil {
			results <- fmt.Sprintf("Player %s CRASHED: %v", playerID, r)
		}
	}()
	
	for i := 0; i < moves; i++ {
		// Random delay
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		
		// Random position
		row := rand.Intn(board.rows)
		col := rand.Intn(board.cols)
		
		err := board.Flip(playerID, row, col)
		if err != nil {
			// Expected errors are ok (empty space, controlled card, etc)
			continue
		}
	}
	
	results <- fmt.Sprintf("Player %s completed successfully", playerID)
}

func runSimulation(numPlayers, movesPerPlayer int) {
	fmt.Printf("Starting simulation: %d players, %d moves each\n", numPlayers, movesPerPlayer)
	
	// Create board
	board, err := ParseFromFile("boards/perfect.txt")
	if err != nil {
		fmt.Printf("Failed to load board: %v\n", err)
		return
	}
	
	results := make(chan string, numPlayers)
	
	// Start all players
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("player%d", i)
		go simulatePlayer(board, playerID, movesPerPlayer, results)
	}
	
	// Wait for all players
	crashes := 0
	for i := 0; i < numPlayers; i++ {
		result := <-results
		fmt.Println(result)
		if len(result) > 0 && result[len(result)-7:] == "CRASHED" {
			crashes++
		}
	}
	
	if crashes == 0 {
		fmt.Println("✓ Simulation completed successfully - no crashes!")
	} else {
		fmt.Printf("✗ %d players crashed\n", crashes)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	fmt.Println("=== Memory Scramble Simulation ===\n")
	
	// Test 1: Few players, many moves
	runSimulation(3, 50)
	fmt.Println()
	
	// Test 2: Many players, few moves
	runSimulation(10, 20)
	fmt.Println()
	
	// Test 3: Stress test
	runSimulation(20, 30)
}