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

type Board struct {
	cells [9]Mark
}

func NewBoard() Board {
	return Board{}
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
	return Empty, false
}

func (b Board) Full() bool {
	return false
}
