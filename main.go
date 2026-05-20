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
	SnakeSymbol = '▣' // U+2588
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
		Position: Position{X: 0, Y: 0},
		Symbol:   SnakeSymbol,
		Size:     1,
	}
	err = board.SetCell(snake.Position, SnakeSymbol)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// Main game loop
	for {
		start := time.Now()
		if gameover {
			print("GAME OVER")
			print(ShowCursor)
			return
		}
		userQuit := handleInput(inputChan, board, snake, log)
		if userQuit {
			return
		}
		err := render(layout, board, help, log)
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

// render the board
func render(layout *Layout, board *Board, help *Help, log *Log) (err error) {
	// use buffer to store board updates
	writer := bufio.NewWriter(os.Stdout)

	// flush entire screen update once at func return
	defer func() {
		if flushErr := writer.Flush(); err == nil {
			err = flushErr
		}
	}()

	err = board.Render(layout, writer)
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
	if pos.Y < 0 || pos.Y >= len(board.grid) {
		return errors.New("position out of bounds (Y)")
	}
	// out of bounds X
	if pos.X < 0 || pos.X >= len(board.grid[pos.Y]) {
		return errors.New("position out of bounds (X)")
	}

	// safe to update board cell
	board.grid[pos.Y][pos.X] = symbol
	return nil
}

func moveSnakeRight(b *Board, s *Snake) error {
	prevPosition := s.Position
	newPosition := Position{X: prevPosition.X + 1, Y: prevPosition.Y}
	s.Position = newPosition

	// move snake
	err := b.SetCell(newPosition, s.Symbol)
	if err != nil {
		return err
	}

	// reset old snake position
	err = b.SetCell(prevPosition, BoardSymbol)
	if err != nil {
		return err
	}

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
func handleInput(inputChan chan string, b *Board, s *Snake, log *Log) bool {
	select {
	case key, ok := <-inputChan:
		if !ok {
			return true
		}
		log.Content = fmt.Sprintf("last key pressed: %s", key)
		switch key {
		case "q":
			return true
		case "l":
			err := moveSnakeRight(b, s)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return false
			}
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
