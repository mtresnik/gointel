package gointel

import "github.com/mtresnik/goutils/pkg/goutils"

type CSPAgent[VAR comparable, DOMAIN comparable] struct {
	SortedVariables   []VAR
	DomainMap         map[VAR][]DOMAIN
	Stack             []CSPNode[VAR, DOMAIN]
	localConstraints  map[VAR][]LocalConstraint[VAR, DOMAIN]
	globalConstraints []GlobalConstraint[VAR, DOMAIN]
}

func NewCSPAgent[VAR comparable, DOMAIN comparable](domainMap map[VAR][]DOMAIN, sortedVariables []VAR, stack []CSPNode[VAR, DOMAIN]) *CSPAgent[VAR, DOMAIN] {
	return &CSPAgent[VAR, DOMAIN]{
		SortedVariables: sortedVariables,
		DomainMap:       domainMap,
		Stack:           stack,
	}
}

func (C *CSPAgent[VAR, DOMAIN]) AddConstraint(constraint Constraint[VAR, DOMAIN]) {
	local := constraint.AsLocal()
	if local != nil {
		for _, variable := range (*local).GetVariables() {
			variableIndex := goutils.IndexOf(C.SortedVariables, func(v VAR) bool {
				return v == variable
			})
			if variableIndex != -1 {
				constraintLookup := []LocalConstraint[VAR, DOMAIN]{}
				temp, ok := C.localConstraints[variable]
				if ok {
					constraintLookup = temp
				}
				constraintLookup = append(constraintLookup, *local)
				C.localConstraints[variable] = constraintLookup
			}
		}
	} else {
		C.globalConstraints = append(C.globalConstraints, constraint)
	}
}

func (C *CSPAgent[VAR, DOMAIN]) AddAllConstraints(constraints ...Constraint[VAR, DOMAIN]) {
	for _, constraint := range constraints {
		C.AddConstraint(constraint)
	}
}

func (C *CSPAgent[VAR, DOMAIN]) GenerateSolutionChannel() chan map[VAR]DOMAIN {
	ch := make(chan map[VAR]DOMAIN)
	go func() {
		defer close(ch)
		var current CSPNode[VAR, DOMAIN]
		for len(C.Stack) > 0 {
			current = C.Stack[len(C.Stack)-1]
			C.Stack = C.Stack[:len(C.Stack)-1]
			currentMap := current.Map
			if IsLocallyConsistent(current.Variable, currentMap, C.GetLocalConstraints(), C.GetGlobalConstraints()) {
				if len(currentMap) == len(C.SortedVariables) {
					if IsConsistent(current.Variable, currentMap, C.GetLocalConstraints(), C.GetGlobalConstraints()) {
						ch <- currentMap
					}
					continue
				} else {
					nextVariableIndex := goutils.IndexOf(C.SortedVariables, func(variable VAR) bool {
						_, ok := currentMap[variable]
						return !ok
					})
					if nextVariableIndex == -1 {
						continue
					}
					nextVariable := C.SortedVariables[nextVariableIndex]
					nextDomain, ok := C.DomainMap[nextVariable]
					if !ok {
						continue
					}
					for _, domain := range nextDomain {
						C.Stack = append(C.Stack, *NewCSPNode(nextVariable, domain, &current))
					}
				}
			}
		}
	}()
	return ch
}

func (C *CSPAgent[VAR, DOMAIN]) Preprocess() {
	// Pass
}

func (C *CSPAgent[VAR, DOMAIN]) GetLocalConstraints() map[VAR][]LocalConstraint[VAR, DOMAIN] {
	return C.localConstraints
}

func (C *CSPAgent[VAR, DOMAIN]) FindAllSolutions() []map[VAR]DOMAIN {
	collected := []map[VAR]DOMAIN{}
	for solution := range C.GenerateSolutionChannel() {
		collected = append(collected, solution)
	}
	return collected
}

func (C *CSPAgent[VAR, DOMAIN]) FindOneSolution() map[VAR]DOMAIN {
	for solution := range C.GenerateSolutionChannel() {
		if solution != nil {
			// Early escape
			return solution
		}
	}
	return nil
}

func (C *CSPAgent[VAR, DOMAIN]) GetDomainMap() map[VAR][]DOMAIN {
	return C.DomainMap
}

func (C *CSPAgent[VAR, DOMAIN]) SetDomainMap(m map[VAR][]DOMAIN) {
	C.DomainMap = m
}

func (C *CSPAgent[VAR, DOMAIN]) Variables() []VAR {
	return C.SortedVariables
}

func (C *CSPAgent[VAR, DOMAIN]) Domains(variable VAR) []DOMAIN {
	return C.DomainMap[variable]
}

func (C *CSPAgent[VAR, DOMAIN]) Contains(v VAR) bool {
	_, ok := C.DomainMap[v]
	return ok
}

func (C *CSPAgent[VAR, DOMAIN]) GetGlobalConstraints() []GlobalConstraint[VAR, DOMAIN] {
	return C.globalConstraints
}
