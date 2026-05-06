package main

import (
	"fmt"
	"time"
)

type Board struct {
	grid [10][10]rune
}

type Position struct {
	X, Y int
}

// For now only position will need to sort out how to add length
type Snake struct {
	Position
	Symbol rune
}

func main() {
	gameover := false

	snake := Snake{Symbol: '★',
		Position: Position{X: 0, Y: 0},
	}

	board := NewBoard()
	Render(board)

	// Main game loop
	for {
		if gameover {
			print("GAME OVER")
			break
		}
		fmt.Print("\033[10A")
		board.grid[snake.Y][snake.X] = snake.Symbol
		Render(board)
		time.Sleep(200 * time.Millisecond)
		if snake.X >= len(board.grid)-1 {
			break
		}
		board.grid[snake.Y][snake.X] = '·'
		snake.X++
	}
}

// main render function
func Render(b *Board) {
	for _, row := range b.grid {
		for _, val := range row {
			fmt.Printf("%2c", val)
		}
		fmt.Println()
	}
}

func NewBoard() *Board {
	b := &Board{}
	for i := range 10 {
		for j := range 10 {
			b.grid[j][i] = '·'
		}
	}
	return b
}
