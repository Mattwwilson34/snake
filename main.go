package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	Rows        = 10
	Cols        = 10
	CursorHome  = "\033[H"
	ClearScreen = "\033[2J"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
	FullBlock   = '█' // U+2588
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

	// Switch to raw mode (Unix/macOS) to allow instant stdin
	err := exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	// TODO: Defer function error handling is confusing review concept
	defer func() {
		err := exec.Command("stty", "-F", "/dev/tty", "-raw", "echo").Run()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()
	fmt.Print(ClearScreen)
	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)

	inputChan := make(chan string)
	go func() {
		for {
			b := make([]byte, 1)
			_, err := os.Stdin.Read(b)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
			inputChan <- string(b)
		}
	}()

	gameover := false
	targetFrameTime := 16 * time.Millisecond

	board := newBoard()
	snake := Snake{Position: Position{X: 5, Y: 5}, Symbol: FullBlock}
	err = setCell(snake.Position, board, snake.Symbol)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	err = render(board)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Main game loop
	for {
		start := time.Now()
		select {
		case key := <-inputChan:
			fmt.Print(key)
			if key == "q" {
				fmt.Print("\r\nQ was pressed. Exiting...\r\n")
				return
			}
		default:
			if gameover {
				print("GAME OVER")
				print(ShowCursor)
				return
			}
			err := render(board)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				break
			}
			elapsed := time.Since(start)
			if elapsed < targetFrameTime {
				time.Sleep(targetFrameTime - elapsed)
			}
		}
	}
}

// render the board
func render(b *Board) error {
	// use buffer to store board updates
	writer := bufio.NewWriter(os.Stdout)

	// move cursor to (0,0) of termianal
	_, err := fmt.Fprint(writer, CursorHome)
	if err != nil {
		return err
	}

	// update board
	for _, row := range b.grid {
		for _, val := range row {
			_, err = fmt.Fprintf(writer, "%2c", val)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprint(writer, "\r\n")
		if err != nil {
			return err
		}
	}

	// flush buffer to stdout in single call to prevent flicker
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

// initialize an empty board
func newBoard() *Board {
	b := &Board{}
	for i := range len(b.grid) {
		for j := range len(b.grid[0]) {
			b.grid[j][i] = '·'
		}
	}
	return b
}

// set the value of a board cell
func setCell(position Position, b *Board, symbol rune) error {
	height := len(b.grid)
	width := len(b.grid[0])

	outOfBoundX := position.X < 0 || position.X >= width
	outOfBoundY := position.Y < 0 || position.Y >= height

	if outOfBoundX || outOfBoundY {
		return errors.New("position out of bounds")
	}

	b.grid[position.Y][position.X] = symbol

	return nil
}
