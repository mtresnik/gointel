package gointel

import "testing"

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
