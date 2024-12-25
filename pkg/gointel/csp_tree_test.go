package gointel

import "testing"

func TestCSPTree_FindAllSolutions(t *testing.T) {
	domainA := []int{3, 4, 5, 6}
	domainB := []int{3, 4}
	domainC := []int{2, 3, 4, 5}
	domainD := []int{2, 3, 4}
	domainE := []int{3, 4}
	domainMap := map[string][]int{
		"A": domainA,
		"B": domainB,
		"C": domainC,
		"D": domainD,
		"E": domainE,
	}
	csp := NewCSPTree(domainMap)
	constraint := GlobalAllDifferentConstraint[string, int]{}
	var c Constraint[string, int] = &constraint
	csp.AddConstraint(&c)
	solutions := csp.FindAllSolutions()
	for _, solution := range solutions {
		for k, v := range solution {
			println(k, ":", v)
		}
		println("------------------")
	}
}
