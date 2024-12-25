package gointel

import (
	"fmt"
	"github.com/mtresnik/goutils/pkg/goutils"
	"math/rand"
	"testing"
	"time"
)

func GenerateValidSudoku(prefilled [][]int) [][]int {
	rand.Seed(time.Now().UnixNano())

	board := make([][]int, 9)
	for i := range board {
		board[i] = make([]int, 9)
		if prefilled != nil {
			copy(board[i], prefilled[i])
		}
	}

	rowUsed := make([]map[int]bool, 9)
	colUsed := make([]map[int]bool, 9)
	boxUsed := make([]map[int]bool, 9)
	for i := 0; i < 9; i++ {
		rowUsed[i] = make(map[int]bool)
		colUsed[i] = make(map[int]bool)
		boxUsed[i] = make(map[int]bool)
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if board[row][col] != 0 {
				num := board[row][col]
				box := (row/3)*3 + col/3
				rowUsed[row][num] = true
				colUsed[col][num] = true
				boxUsed[box][num] = true
			}
		}
	}

	boxIndex := func(row, col int) int {
		return (row/3)*3 + col/3
	}

	type position struct {
		row, col, value int
	}
	stack := []position{}

	row, col := 0, 0

	for row < 9 {
		if prefilled != nil && prefilled[row][col] != 0 {
			col++
			if col == 9 {
				row++
				col = 0
			}
			continue
		}

		found := false
		for num := board[row][col] + 1; num <= 9; num++ {
			box := boxIndex(row, col)
			if !rowUsed[row][num] && !colUsed[col][num] && !boxUsed[box][num] {
				board[row][col] = num
				rowUsed[row][num] = true
				colUsed[col][num] = true
				boxUsed[box][num] = true

				stack = append(stack, position{row, col, num})

				col++
				if col == 9 {
					row++
					col = 0
				}
				found = true
				break
			}
		}

		if !found {
			board[row][col] = 0

			if len(stack) == 0 {
				return nil
			}

			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			box := boxIndex(last.row, last.col)
			delete(rowUsed[last.row], last.value)
			delete(colUsed[last.col], last.value)
			delete(boxUsed[box], last.value)

			row, col = last.row, last.col
		}
	}

	return board
}

func TestCSPDomain_Sudoku(t *testing.T) {
	prefilled := make([][]int, 9)
	for i := range prefilled {
		prefilled[i] = make([]int, 9)
	}
	shuffled := goutils.RangeOfInts(1, 10)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	for i := 0; i < 9; i++ {
		prefilled[i][i] = shuffled[i]
	}

	board := GenerateValidSudoku(prefilled)

	fmt.Println("Generated Valid Sudoku Board:")
	for _, row := range board {
		fmt.Println(row)
	}
}
