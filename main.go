package main

import (
	"errors"
	"fmt"
	"time"
)

const (
	Rows = 10
	Cols = 10
)

type Board struct {
	grid [Cols][Rows]rune
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
		err := SetCell(Position{X: snake.X, Y: 0}, board, snake.Symbol)
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		Render(board)
		time.Sleep(200 * time.Millisecond)
		board.grid[snake.Y][snake.X] = '·'
		snake.X++
	}
}

// render the board
func Render(b *Board) {
	for _, row := range b.grid {
		for _, val := range row {
			fmt.Printf("%2c", val)
		}
		fmt.Println()
	}
}

// initialize an empty board
func NewBoard() *Board {
	b := &Board{}
	for i := range len(b.grid) {
		for j := range len(b.grid[0]) {
			b.grid[j][i] = '·'
		}
	}
	return b
}

// set the value of a board cell
func SetCell(position Position, b *Board, symbol rune) error {
	height := len(b.grid)
	width := len(b.grid[0])

	outOfBoundX := position.X < 0 || position.X > width
	outOfBoundY := position.Y < 0 || position.Y > height

	if outOfBoundX || outOfBoundY {
		return errors.New("position is out of bound")
	}

	b.grid[position.Y][position.X] = symbol

	return nil
}
