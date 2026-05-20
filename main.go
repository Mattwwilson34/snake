package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	SmallBoard  = 15
	MediumBoard = 30
	LargeBoard  = 45
)

const (
	// terminal operators
	CursorHome  = "\033[H"
	ClearScreen = "\033[2J"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
	// symbols
	SnakeSymbol = '▣'
	BoardSymbol = '·'
)

type Board struct {
	grid [][]rune
}
type Help struct {
	Content string
}
type Log struct {
	Content string
}

type Position struct {
	X, Y int
}

// For now only position will need to sort out how to add length
type Snake struct {
	Position
	Symbol rune
	Size   int
}

func main() {
	// context to kill go routines on exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// prompt user for board size
	boardSize, err := promptForBoardSize()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// current board is square can update later to accommodate new shapes
	layout := NewLayout(boardSize, boardSize)
	fmt.Printf("%#v\n", layout)

	help := &Help{
		Content: "[q] Quit",
	}
	log := &Log{}
	board := newBoard(boardSize)

	// setup terminal for game
	cleanup, _ := setupTerminal()
	defer cleanup()

	// start go routine to listen for user input
	inputChan := make(chan string)
	startInputListener(ctx, inputChan)

	gameover := false
	targetFrameTime := 16 * time.Millisecond

	snake := &Snake{
		Position: Position{X: 1, Y: 1},
		Symbol:   SnakeSymbol,
		Size:     1,
	}

	// Main game loop
	for {
		start := time.Now()
		if gameover {
			print("GAME OVER")
			print(ShowCursor)
			return
		}
		userQuit := handleInput(inputChan, snake, log)
		if userQuit {
			return
		}
		err := render(layout, board, snake, help, log)
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

func (help *Help) Render(layout *Layout, writer *bufio.Writer) error {
	// move cursor to start of help region
	cursorErr := MoveCursor(layout.HelpRegion.X, layout.HelpRegion.Y, writer)
	if cursorErr != nil {
		return cursorErr
	}

	_, printErr := fmt.Fprint(writer, help.Content)
	if printErr != nil {
		return printErr
	}
	return nil
}

func (log *Log) Render(layout *Layout, writer *bufio.Writer) error {
	// move cursor to start of log region
	cursorErr := MoveCursor(layout.LogRegion.X, layout.LogRegion.Y, writer)
	if cursorErr != nil {
		return cursorErr
	}

	_, printErr := fmt.Fprint(writer, log.Content)
	if printErr != nil {
		return printErr
	}
	return nil
}

func (snake *Snake) Render(writer *bufio.Writer) error {
	// move cursor to snake position
	cursorErr := MoveCursor(snake.X, snake.Y, writer)
	if cursorErr != nil {
		return cursorErr
	}
	_, writeErr := fmt.Fprintf(writer, "%2c", snake.Symbol)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func (board *Board) Render(layout *Layout, writer *bufio.Writer) error {
	// move cursor to start of board region
	cursorErr := MoveCursor(layout.BoardRegion.X, layout.BoardRegion.Y, writer)
	if cursorErr != nil {
		return cursorErr
	}

	// write board to buffer
	for _, row := range board.grid {
		for _, val := range row {
			_, err := fmt.Fprintf(writer, "%2c", val)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(writer, "\r\n")
		if err != nil {
			return err
		}
	}

	return nil

}

// main render function
func render(layout *Layout, board *Board, snake *Snake, help *Help, log *Log) (err error) {
	// use buffer to store board updates
	writer := bufio.NewWriter(os.Stdout)

	// flush entire screen update once at func return
	defer func() {
		if flushErr := writer.Flush(); err == nil {
			err = flushErr
		}
	}()

	log.Content = fmt.Sprintf("%#v\n", snake)
	err = board.Render(layout, writer)
	if err != nil {
		return err
	}
	err = snake.Render(writer)
	if err != nil {
		return err
	}
	err = help.Render(layout, writer)
	if err != nil {
		return err
	}
	err = log.Render(layout, writer)
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
		for j := range len(b.grid[i]) {
			b.grid[i][j] = BoardSymbol
		}
	}
	return b
}

// set the value of a board cell
func (board *Board) SetCell(pos Position, symbol rune) error {

	// handle no board
	if len(board.grid) == 0 {
		return errors.New("board uninitialized")
	}
	// out of bounds Y
	if board.IsOutOfBounds(pos.X, pos.Y) {
		return errors.New("position out of bounds")
	}

	// safe to update board cell
	board.grid[pos.Y][pos.X] = symbol
	return nil
}

func (board *Board) IsOutOfBounds(x, y int) bool {
	// out of bounds Y
	if y < 0 || y >= len(board.grid) {
		return true
	}
	// out of bounds X
	if x < 0 || x >= len(board.grid[y]) {
		return true
	}
	return false
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

// handle user input
func handleInput(inputChan chan string, s *Snake, log *Log) bool {
	select {
	case key, ok := <-inputChan:
		if !ok {
			return true
		}
		log.Content = fmt.Sprintf("last key pressed: %s", key)
		switch key {

		// quit game
		case "q":
			return true
		// snake movements
		// up
		case "k":
			s.Y = s.Y - 1
		// down
		case "j":
			s.Y = s.Y + 1
		// left
		case "h":
			s.X = s.X - 2
		// right
		case "l":
			s.X = s.X + 2
		}
	default:
		return false
	}
	return false
}

// parse input into numerical board size
func parseBoardSize(input string) (int, bool) {
	switch input {
	case "s", "small", "1":
		return SmallBoard, true
	case "m", "medium", "2":
		return MediumBoard, true
	case "l", "large", "3":
		return LargeBoard, true
	default:
		return 0, false
	}
}

// prompt the user for board size
func promptForBoardSize() (int, error) {
	fmt.Println("Select a board size.")
	fmt.Println("1. (s)mall, 2. (m)edium, 3. (l)arge")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	cleanInput := strings.ToLower(strings.TrimSpace(input))
	boardSize, success := parseBoardSize(cleanInput)
	if !success {
		return 0, errors.New("failed to parse board size from user input")
	}
	return boardSize, nil

}

// move cursor to x,y within buffer
func MoveCursor(x, y int, writer *bufio.Writer) error {
	// \033[y;xH tells the terminal to move to row y, column x
	_, cursorErr := fmt.Fprintf(writer, "\033[%d;%dH", y, x)
	if cursorErr != nil {
		return cursorErr
	}
	return nil
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
