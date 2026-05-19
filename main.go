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

type BoardSize int

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

type Position struct {
	X, Y int
}

// For now only position will need to sort out how to add length
type Snake struct {
	Position
	Symbol rune
	Size   int
}

type Message struct {
	Message     string
	Expiraton   int
	DisplayTime int
}

type GameInfo struct {
	Message        Message
	LastKeyPressed string
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

	gameInfo := &GameInfo{
		Message: Message{
			Message:     "",
			Expiraton:   3600,
			DisplayTime: 0,
		},
		LastKeyPressed: "",
	}

	// Main game loop
	for {
		start := time.Now()
		if gameover {
			print("GAME OVER")
			print(ShowCursor)
			return
		}
		userQuit := handleInput(inputChan, board, snake, gameInfo)
		if userQuit {
			return
		}
		err := render(board, gameInfo)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}
		elapsed := time.Since(start)
		if elapsed < targetFrameTime {
			time.Sleep(targetFrameTime - elapsed)
		}
		if gameInfo.Message.Message != "" {
			gameInfo.Message.DisplayTime++
		}
		// reset user message if it has expired
		if gameInfo.Message.DisplayTime >= gameInfo.Message.Expiraton {
			gameInfo.Message.Message = ""
			gameInfo.Message.DisplayTime = 0
		}
	}
}

// render the board
func render(b *Board, gameInfo *GameInfo) error {
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

	// print any user message
	_, err = fmt.Fprintf(writer, "message: %v\r\n", gameInfo.Message.Message)
	if err != nil {
		return err
	}

	// print last key pressed
	_, err = fmt.Fprintf(writer, "last key pressed: %v", gameInfo.LastKeyPressed)
	if err != nil {
		return err
	}

	// flush buffer to stdout in single call to prevent flicker
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

// initialize an empty board
func newBoard(size BoardSize) *Board {
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
func (b *Board) SetCell(pos Position, symbol rune) error {

	// handle no board
	if len(b.grid) == 0 {
		return errors.New("board uninitialized")
	}
	// out of bounds Y
	if pos.Y < 0 || pos.Y >= len(b.grid) {
		return errors.New("position out of bounds (Y)")
	}
	// out of bounds X
	if pos.X < 0 || pos.X >= len(b.grid[pos.Y]) {
		return errors.New("position out of bounds (X)")
	}

	// safe to update board cell
	b.grid[pos.Y][pos.X] = symbol
	return nil
}

func moveSnakeRight(b *Board, s *Snake, g *GameInfo) error {
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
func handleInput(inputChan chan string, b *Board, s *Snake, g *GameInfo) bool {
	select {
	case key, ok := <-inputChan:
		if !ok {
			return true
		}
		fmt.Println(key)
		switch key {
		case "q":
			return true
		case "l":
			err := moveSnakeRight(b, s, g)
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
func parseBoardSize(input string) (BoardSize, bool) {
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
func promptForBoardSize() (BoardSize, error) {
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
