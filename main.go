package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Welcome to SNAKE")

	gameover := false
	board := [10][10]int{}
	snake := 4
	snakeX := 1
	snakeY := 1

	render(&board)

	// Main game loop
	for {
		if gameover {
			print("GAME OVER")
			break
		}
		fmt.Print("\033[10A")
		board[snakeY][snakeX] = snake
		render(&board)
		time.Sleep(200 * time.Millisecond)
		if snakeX >= len(board)-1 {
			break
		}
		board[snakeY][snakeX] = 0
		snakeX++
	}
}

// main render function
func render(board *[10][10]int) {
	for _, row := range board {
		for _, val := range row {
			fmt.Printf("%2d", val)
		}
		fmt.Println()
	}
}
