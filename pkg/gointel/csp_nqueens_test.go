package gointel

import (
	"fmt"
	"github.com/mtresnik/goutils/pkg/goutils"
	"image"
	"image/png"
	"os"
	"runtime"
	"sync"
	"testing"
)

func isSafe(board []int, row, col int) bool {
	for i := 0; i < row; i++ {
		if board[i] == col || intAbs(board[i]-col) == intAbs(i-row) {
			return false
		}
	}
	return true
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func solveNQueens(n int) chan []int {
	solutions := make(chan []int, 1000)
	if n <= 0 {
		close(solutions)
		return solutions
	}
	if n == 1 {
		solutions <- []int{0}
		close(solutions)
		return solutions
	}
	if n < 4 {
		close(solutions)
		return solutions
	}

	chunks := runtime.NumCPU()
	wg := sync.WaitGroup{}

	maxFirstCol := n / 2
	if n%2 == 1 {
		maxFirstCol = n/2 + 1
	}

	for i := 0; i < chunks; i++ {
		start := i * maxFirstCol / chunks
		end := (i + 1) * maxFirstCol / chunks

		wg.Add(1)
		go func(startCol, endCol int) {
			defer wg.Done()
			board := make([]int, n)
			for i := range board {
				board[i] = -1
			}

			for col := startCol; col < endCol; col++ {
				board[0] = col
				stack := []int{1}
				currentCol := 0

				for len(stack) > 0 {
					row := len(stack)
					if currentCol >= n {
						stack = stack[:len(stack)-1]
						if len(stack) > 0 {
							currentCol = board[len(stack)] + 1
						}
						continue
					}

					if isSafe(board, row, currentCol) {
						board[row] = currentCol
						if row == n-1 {
							solution := make([]int, n)
							copy(solution, board)
							solutions <- solution

							if col < n/2 {
								mirror := make([]int, n)
								for i := 0; i < n; i++ {
									mirror[i] = n - 1 - board[i]
								}
								solutions <- mirror
							} else if n%2 == 1 && col == n/2 {
								currentCol++
								continue
							}
							currentCol++
						} else {
							stack = append(stack, 0)
							currentCol = 0
						}
					} else {
						currentCol++
					}
				}
			}
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(solutions)
	}()

	return solutions
}

func TestNQueens(t *testing.T) {
	for n := 0; n < 13; n++ {

		squareSize := 20
		imageSize := squareSize * n
		img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
		goutils.FillRectangle(img, 0, 0, imageSize, imageSize, goutils.COLOR_WHITE)
		frequency := make([][]float64, n)
		for i := 0; i < n; i++ {
			frequency[i] = make([]float64, n)
		}
		count := 0
		lastSolution := []int{}
		for solution := range solveNQueens(n) {
			for col, row := range solution {
				frequency[col][row]++
			}
			count++
			lastSolution = solution
		}
		fmt.Printf("N=%d, Count=%d\n", n, count)
		if count == 0 {
			continue
		}

		lastImage := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
		goutils.FillRectangle(img, 0, 0, imageSize, imageSize, goutils.COLOR_WHITE)
		for imageRow := 0; imageRow < imageSize; imageRow++ {
			row := imageRow / squareSize
			for imageCol := 0; imageCol < imageSize; imageCol++ {
				col := imageCol / squareSize
				color := goutils.COLOR_WHITE
				if (row+col)%2 == 0 {
					color = goutils.COLOR_BLACK
				}
				goutils.FillRectangle(lastImage, imageCol, imageRow, squareSize, squareSize, color)
			}
		}
		for col, row := range lastSolution {
			imageRow := row * squareSize
			imageCol := col * squareSize
			centerCol := imageCol + squareSize/2
			centerRow := imageRow + squareSize/2
			goutils.FillCircle(lastImage, centerCol, centerRow, squareSize/2, goutils.COLOR_RED)
		}
		{
			lastImageName := fmt.Sprintf("TestNQueens_OneSolution_%d.png", n)
			file, err := os.Create(lastImageName)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			png.Encode(file, lastImage)
		}

		// Scale frequency
		minFrequency := frequency[0][0]
		maxFrequency := frequency[0][0]
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if frequency[i][j] < minFrequency {
					minFrequency = frequency[i][j]
				}
				if frequency[i][j] > maxFrequency {
					maxFrequency = frequency[i][j]
				}
			}
		}
		delta := maxFrequency - minFrequency
		scaled := make([][]float64, n)
		for i := 0; i < n; i++ {
			scaled[i] = make([]float64, n)
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				scaled[i][j] = (frequency[i][j] - minFrequency) / delta
			}
		}
		for imageRow := 0; imageRow < imageSize; imageRow++ {
			row := imageRow / squareSize
			for imageCol := 0; imageCol < imageSize; imageCol++ {
				col := imageCol / squareSize
				scaledValue := scaled[col][row]
				if scaledValue > 1.0 {
					scaledValue = 1.0
				}
				if scaledValue < 0.0 {
					scaledValue = 0.0
				}
				color := goutils.Gradient(scaledValue, goutils.COLOR_WHITE, goutils.COLOR_BLACK)
				goutils.FillRectangle(img, imageCol, imageRow, squareSize, squareSize, color)
			}
		}
		imageName := fmt.Sprintf("TestNQueens_%d.png", n)
		file, err := os.Create(imageName)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		png.Encode(file, img)
	}
}
