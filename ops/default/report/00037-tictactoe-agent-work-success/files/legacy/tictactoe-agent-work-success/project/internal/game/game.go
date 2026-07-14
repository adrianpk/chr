package game

import (
	"fmt"
	"strconv"
	"strings"
)

type Mark string

const (
	Empty Mark = ""
	X     Mark = "X"
	O     Mark = "O"
)

type Outcome int

const (
	OutcomeInProgress Outcome = iota
	OutcomeXWon
	OutcomeOWon
	OutcomeDraw
)

type Board struct {
	cells [9]Mark
}

type Game struct {
	board         Board
	currentPlayer Mark
	outcome       Outcome
}

func NewBoard() Board {
	return Board{}
}

func NewGame() Game {
	return Game{
		board:         NewBoard(),
		currentPlayer: X,
		outcome:       OutcomeInProgress,
	}
}

func ParsePosition(input string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0, fmt.Errorf("position must be a number from 1 to 9")
	}

	if value < 1 || value > 9 {
		return 0, fmt.Errorf("position must be between 1 and 9")
	}

	return value - 1, nil
}

func (g Game) Board() Board {
	return g.board
}

func (g Game) CurrentPlayer() Mark {
	return g.currentPlayer
}

func (g Game) Outcome() Outcome {
	return g.outcome
}

func (g Game) Over() bool {
	return g.outcome != OutcomeInProgress
}

func (g Game) Draw() bool {
	return g.outcome == OutcomeDraw
}

func (g Game) Winner() (Mark, bool) {
	switch g.outcome {
	case OutcomeXWon:
		return X, true
	case OutcomeOWon:
		return O, true
	default:
		return Empty, false
	}
}

func (g Game) Play(index int) (Game, error) {
	return g.ApplyMove(index, g.currentPlayer)
}

func (g Game) ApplyMove(index int, mark Mark) (Game, error) {
	if g.Over() {
		return g, fmt.Errorf("game is already over")
	}

	if mark != X && mark != O {
		return g, fmt.Errorf("mark must be X or O")
	}

	if mark != g.currentPlayer {
		return g, fmt.Errorf("it is %s's turn", g.currentPlayer)
	}

	board, err := g.board.Place(index, mark)
	if err != nil {
		return g, err
	}

	next := Game{
		board:         board,
		currentPlayer: mark,
		outcome:       outcomeForBoard(board),
	}
	if next.outcome == OutcomeInProgress {
		next.currentPlayer = nextPlayer(mark)
	}

	return next, nil
}

func (b Board) At(index int) (Mark, error) {
	if index < 0 || index >= len(b.cells) {
		return Empty, fmt.Errorf("cell index out of range")
	}

	return b.cells[index], nil
}

func (b Board) Place(index int, mark Mark) (Board, error) {
	if mark != X && mark != O {
		return b, fmt.Errorf("mark must be X or O")
	}

	if index < 0 || index >= len(b.cells) {
		return b, fmt.Errorf("cell index out of range")
	}

	if b.cells[index] != Empty {
		return b, fmt.Errorf("cell is already occupied")
	}

	b.cells[index] = mark

	return b, nil
}

func (b Board) Winner() (Mark, bool) {
	lines := [][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for _, line := range lines {
		mark := b.cells[line[0]]
		if mark != Empty && mark == b.cells[line[1]] && mark == b.cells[line[2]] {
			return mark, true
		}
	}

	return Empty, false
}

func (b Board) Full() bool {
	for _, cell := range b.cells {
		if cell == Empty {
			return false
		}
	}

	return true
}

func outcomeForBoard(board Board) Outcome {
	winner, ok := board.Winner()
	if ok {
		switch winner {
		case X:
			return OutcomeXWon
		case O:
			return OutcomeOWon
		}
	}

	if board.Full() {
		return OutcomeDraw
	}

	return OutcomeInProgress
}

func nextPlayer(mark Mark) Mark {
	if mark == X {
		return O
	}

	return X
}
