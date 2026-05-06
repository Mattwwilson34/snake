package main

import (
	"fmt"
	"time"
)

type Position struct {
	X, Y int
}

// For now only position will need to sort out how to add length
type Snake struct {
	Position
	Symbol rune
}

func main() {
	fmt.Println("Welcome to SNAKE")

	gameover := false

	board := [10][10]rune{}
	dot := '\u00B7'
	for i := range 10 {
		for j := range 10 {
			board[j][i] = dot
		}
	}
	snake := Snake{Symbol: '★',
		Position: Position{X: 0, Y: 0},
	}

	render(&board)

	// Main game loop
	for {
		if gameover {
			print("GAME OVER")
			break
		}
		fmt.Print("\033[10A")
		board[snake.Y][snake.X] = snake.Symbol
		render(&board)
		time.Sleep(200 * time.Millisecond)
		if snake.X >= len(board)-1 {
			break
		}
		board[snake.Y][snake.X] = dot
		snake.X++
	}
}

// main render function
func render(board *[10][10]rune) {
	for _, row := range board {
		for _, val := range row {
			fmt.Printf("%2c", val)
		}
		fmt.Println()
	}
}
