package game

import (
	"errors"
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

// Rejection categories let callers classify errors with errors.Is.
var (
	ErrInvalidInput       = errors.New("position must be a number from 1 to 9")
	ErrPositionOutOfRange = errors.New("position must be between 1 and 9")
	ErrIndexOutOfRange    = errors.New("cell index out of range")
	ErrInvalidMark        = errors.New("mark must be X or O")
	ErrCellOccupied       = errors.New("cell is already occupied")
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
		return 0, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}

	if value < 1 || value > 9 {
		return 0, ErrPositionOutOfRange
	}

	return value - 1, nil
}

func (b Board) At(index int) (Mark, error) {
	if index < 0 || index >= len(b.cells) {
		return Empty, ErrIndexOutOfRange
	}

	return b.cells[index], nil
}

func (b Board) Place(index int, mark Mark) (Board, error) {
	if mark != X && mark != O {
		return b, ErrInvalidMark
	}

	if index < 0 || index >= len(b.cells) {
		return b, ErrIndexOutOfRange
	}

	if b.cells[index] != Empty {
		return b, ErrCellOccupied
	}

	b.cells[index] = mark

	return b, nil
}

func (b Board) Winner() (Mark, bool) {
	lines := [8][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
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
	for _, mark := range b.cells {
		if mark == Empty {
			return false
		}
	}

	return true
}
