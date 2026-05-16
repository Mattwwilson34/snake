package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	LargeBoard  = 60
	MediumBoard = 40
	SmallBoard  = 20
	CursorHome  = "\033[H"
	ClearScreen = "\033[2J"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
	FullBlock   = '█' // U+2588
	BoardSymbol = '·'
)

type Board struct {
	grid [][]rune
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
	// context to kill go routines on exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup terminal for game
	cleanup, _ := setupTerminal()
	defer cleanup()

	// start go routine to listen for user input
	inputChan := make(chan string)
	startInputListener(ctx, inputChan)

	gameover := false
	targetFrameTime := 16 * time.Millisecond
	board := newBoard(SmallBoard)

	// Main game loop
	for {
		start := time.Now()
		if gameover {
			print("GAME OVER")
			print(ShowCursor)
			return
		}
		userQuit := handleInput(inputChan)
		if userQuit {
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

// render the board
func render(b *Board) error {
	// use buffer to store board updates
	writer := bufio.NewWriter(os.Stdout)

	// move cursor to (0,0) of terminal
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
func newBoard(size int) *Board {
	b := &Board{
		grid: make([][]rune, size),
	}
	for i := range b.grid {
		b.grid[i] = make([]rune, size)
		for j := range len(b.grid[0]) {
			b.grid[i][j] = BoardSymbol
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

// set terminal to raw, clear screen, and hide cursor
func setupTerminal() (func(), error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New("not a terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	fmt.Print(ClearScreen, HideCursor)

	return func() {
		fmt.Print(ShowCursor)
		_ = term.Restore(fd, oldState)
	}, nil
}

// start a stdin reader in a go routine
func startInputListener(ctx context.Context, inputChan chan string) {
	go func() {
		defer close(inputChan)
		b := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(b)
			if err != nil {
				return
			}

			select {
			case inputChan <- string(b):
			case <-ctx.Done():
				return
			}
		}
	}()
}

// handle user input
func handleInput(inputChan chan string) bool {
	select {
	case key, ok := <-inputChan:
		if !ok {
			return true
		}
		switch key {
		case "q":
			return true
		}
	default:
		return false
	}
	return false
}
