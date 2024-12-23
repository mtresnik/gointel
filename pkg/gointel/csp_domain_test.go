package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
	"image"
	"image/png"
	"math"
	"os"
	"testing"
)

type mapColoringConstraint struct {
	From string
	To   string
}

func (M *mapColoringConstraint) IsSatisfied(assignment map[string]string) bool {
	return true
}

func (M *mapColoringConstraint) AsLocal() *LocalConstraint[string, string] {
	var local LocalConstraint[string, string] = M
	return &local
}

func (M *mapColoringConstraint) IsReusable() bool {
	return false
}

func (M *mapColoringConstraint) IsPossiblySatisfied(assignment map[string]string) bool {
	assignmentFrom, fromOk := assignment[M.From]
	assignmentTo, toOk := assignment[M.To]
	if !fromOk || !toOk {
		return true
	}
	return assignmentFrom != assignmentTo
}

func (M *mapColoringConstraint) GetVariables() []string {
	return []string{M.From, M.To}
}

func TestCSPDomain_FindAllSolutions(t *testing.T) {
	wa := "Western Australia"
	nt := "Northern Territory"
	sa := "South Australia"
	q := "Queensland"
	nsw := "New South Wales"
	v := "Victoria"
	tas := "Tasmania"

	red := "red"
	green := "green"
	blue := "blue"

	colors := []string{red, green, blue}
	variables := []string{wa, nt, sa, q, nsw, v, tas}

	domains := map[string][]string{}
	for _, variable := range variables {
		domains[variable] = colors
	}

	csp := NewCSPDomain(domains)
	constraints := []Constraint[string, string]{
		&mapColoringConstraint{From: wa, To: nt},
		&mapColoringConstraint{From: wa, To: sa},

		&mapColoringConstraint{From: sa, To: nt},

		&mapColoringConstraint{From: q, To: nt},
		&mapColoringConstraint{From: q, To: sa},
		&mapColoringConstraint{From: q, To: nsw},

		&mapColoringConstraint{From: nsw, To: sa},

		&mapColoringConstraint{From: v, To: sa},
		&mapColoringConstraint{From: v, To: nsw},
		&mapColoringConstraint{From: v, To: tas},
	}
	csp.AddAllConstraints(constraints...)
	solutions := csp.FindAllSolutions()

	for _, solution := range solutions {
		for k, v := range solution {
			println(k, ":", v)
		}
		println("------------------")
	}

}

type QueenConstraint struct {
	Columns []int
}

func (q *QueenConstraint) IsSatisfied(_ map[int]int) bool {
	return true
}

func (q *QueenConstraint) AsLocal() *LocalConstraint[int, int] {
	var l LocalConstraint[int, int] = q
	return &l
}

func (q *QueenConstraint) IsReusable() bool {
	return false
}

func (q *QueenConstraint) IsPossiblySatisfied(assignment map[int]int) bool {
	for col1, row1 := range assignment {
		for col2, row2 := range assignment {
			if col1 == col2 {
				continue
			}
			if row1 == row2 {
				return false
			}
			if math.Abs(float64(col1-col2)) == math.Abs(float64(row1-row2)) {
				return false
			}
		}
	}
	return true
}

func (q *QueenConstraint) GetVariables() []int {
	return q.Columns
}

func drawBoard(n int, solutions []map[int]int) *image.RGBA {
	squareSize := 20
	imageSize := squareSize * n
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	goutils.FillRectangle(img, 0, 0, imageSize, imageSize, goutils.COLOR_WHITE)
	frequency := make([][]float64, n)
	for i := 0; i < n; i++ {
		frequency[i] = make([]float64, n)
	}
	for _, solution := range solutions {
		for col, row := range solution {
			frequency[col][row]++
		}
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
	return img
}

func TestCSPDomain_NQueens(t *testing.T) {
	n := 13
	columns := []int{}
	for i := 0; i < n; i++ {
		columns = append(columns, i)
	}
	domains := map[int][]int{}
	for _, column := range columns {
		clonedColumns := make([]int, len(columns))
		copy(clonedColumns, columns)
		domains[column] = clonedColumns
	}
	preprocessor := AC3Preprocessor[int, int]{}
	csp := NewCSPDomain(domains, &preprocessor)
	constraints := []Constraint[int, int]{
		&QueenConstraint{Columns: columns},
	}
	csp.AddAllConstraints(constraints...)
	solutions := csp.FindAllSolutions()
	img := drawBoard(n, solutions)
	file, err := os.Create("TestCSPDomain_NQueens.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}
