package main

import "fmt"

func main() {
	fmt.Println("Welcome to SNAKE")

	board := [10][10]int{}
	render(&board)
}

// main render function
func render(board *[10][10]int) {
	for _, row := range board {
		for _, val := range row {
			fmt.Printf("%2d ", val)
		}
		fmt.Println()
	}
}
