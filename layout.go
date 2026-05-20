package main

type Region struct {
	X, Y, Width, Height int
}

type Layout struct {
	BoardRegion Region
	HelpRegion  Region
	LogRegion   Region
}

func NewLayout(boardWidth, boardHeight int) *Layout {
	// 1. Game board starts at top left
	board := Region{
		X:      0,
		Y:      0,
		Width:  boardWidth,
		Height: boardHeight,
	}

	// 2. Help menu directly below board
	help := Region{
		X:      0,
		Y:      boardHeight + 1, // 1 line padding
		Width:  boardWidth,
		Height: 3,
	}

	// 3. Log output directly below help menu
	log := Region{
		X:      0,
		Y:      help.Y + 1, // 1 line padding
		Width:  boardWidth,
		Height: 3,
	}

	return &Layout{
		BoardRegion: board,
		HelpRegion:  help,
		LogRegion:   log,
	}
}
