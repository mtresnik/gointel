package gointel

import "github.com/mtresnik/goutils/pkg/goutils"

type CSPTree[VAR comparable, DOMAIN comparable] struct {
	DomainMap         map[VAR][]DOMAIN
	Preprocessors     []CSPPreprocessor[VAR, DOMAIN]
	localConstraints  map[VAR][]*LocalConstraint[VAR, DOMAIN]
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]
	sortingFunction   *func(a, b VAR) bool
}

func NewCSPTree[VAR comparable, DOMAIN comparable](domainMap map[VAR][]DOMAIN, preprocessors ...CSPPreprocessor[VAR, DOMAIN]) *CSPTree[VAR, DOMAIN] {
	return &CSPTree[VAR, DOMAIN]{
		DomainMap:         domainMap,
		Preprocessors:     preprocessors,
		localConstraints:  map[VAR][]*LocalConstraint[VAR, DOMAIN]{},
		globalConstraints: []*GlobalConstraint[VAR, DOMAIN]{},
	}
}

func (C *CSPTree[VAR, DOMAIN]) GetDomainMap() map[VAR][]DOMAIN {
	return C.DomainMap
}

func (C *CSPTree[VAR, DOMAIN]) SetDomainMap(m map[VAR][]DOMAIN) {
	C.DomainMap = m
}

func (C *CSPTree[VAR, DOMAIN]) GetVariables() []VAR {
	return goutils.Keys(C.DomainMap)
}

func (C *CSPTree[VAR, DOMAIN]) GetDomainForVariable(variable VAR) []DOMAIN {
	ret, ok := C.DomainMap[variable]
	if !ok {
		return []DOMAIN{}
	}
	return ret
}

func (C *CSPTree[VAR, DOMAIN]) Contains(v VAR) bool {
	_, ok := C.DomainMap[v]
	return ok
}

func (C *CSPTree[VAR, DOMAIN]) Preprocess() {
	var csp CSP[VAR, DOMAIN] = C
	for _, preprocessor := range C.Preprocessors {
		preprocessor.Preprocess(&csp)
	}
}

func (C *CSPTree[VAR, DOMAIN]) GetLocalConstraints() map[VAR][]*LocalConstraint[VAR, DOMAIN] {
	return C.localConstraints
}

func (C *CSPTree[VAR, DOMAIN]) GetGlobalConstraints() []*GlobalConstraint[VAR, DOMAIN] {
	return C.globalConstraints
}

func (C *CSPTree[VAR, DOMAIN]) AddConstraint(constraint *Constraint[VAR, DOMAIN]) {
	if constraint == nil {
		return
	}
	local := (*constraint).AsLocal()
	if local != nil {
		for _, variable := range (*local).GetVariables() {
			variableIndex := goutils.IndexOf(C.GetVariables(), func(v VAR) bool {
				return v == variable
			})
			if variableIndex != -1 {
				constraintLookup := []*LocalConstraint[VAR, DOMAIN]{}
				temp, ok := C.localConstraints[variable]
				if ok {
					constraintLookup = temp
				}
				constraintLookup = append(constraintLookup, local)
				C.localConstraints[variable] = constraintLookup
			}
		}
	} else {
		println("Adding global constraint:")
		var global GlobalConstraint[VAR, DOMAIN] = *constraint
		C.globalConstraints = append(C.globalConstraints, &global)
	}
}

func (C *CSPTree[VAR, DOMAIN]) AddAllConstraints(constraints ...*Constraint[VAR, DOMAIN]) {
	for _, constraint := range constraints {
		C.AddConstraint(constraint)
	}
}

func (C *CSPTree[VAR, DOMAIN]) FindAllSolutions() []map[VAR]DOMAIN {
	C.Preprocess()
	agent := C.constructAgent()
	if agent == nil {
		return []map[VAR]DOMAIN{}
	}
	return agent.FindAllSolutions()
}

func (C *CSPTree[VAR, DOMAIN]) FindOneSolution() map[VAR]DOMAIN {
	C.Preprocess()
	agent := C.constructAgent()
	if agent == nil {
		return nil
	}
	return agent.FindOneSolution()
}

func (C *CSPTree[VAR, DOMAIN]) GenerateSolutionChannel() chan map[VAR]DOMAIN {
	ch := make(chan map[VAR]DOMAIN)
	defer close(ch)
	for _, solution := range C.FindAllSolutions() {
		ch <- solution
	}
	return ch
}

func (C *CSPTree[VAR, DOMAIN]) constructAgent() *CSPAgent[VAR, DOMAIN] {
	variables := C.GetVariables()
	if len(variables) == 0 {
		return nil
	}
	first := variables[0]
	if len(C.DomainMap) == 0 {
		return nil
	}
	firstDomain, ok := C.DomainMap[first]
	if !ok {
		return nil
	}
	rootNodes := []CSPNode[VAR, DOMAIN]{}
	for _, domain := range firstDomain {
		node := NewCSPNode(first, 0, domain, firstDomain)
		if node != nil {
			rootNodes = append(rootNodes, *node)
		}
	}
	ret := NewCSPAgent(&C.DomainMap, variables, rootNodes)
	ret.SortingFunction = C.sortingFunction
	for _, local := range C.localConstraints {
		toAdd := []*Constraint[VAR, DOMAIN]{}
		for _, constraint := range local {
			var c Constraint[VAR, DOMAIN] = *constraint
			toAdd = append(toAdd, &c)
		}
		ret.AddAllConstraints(toAdd...)
	}
	toAdd := []*Constraint[VAR, DOMAIN]{}
	for _, constraint := range C.globalConstraints {
		var c Constraint[VAR, DOMAIN] = *constraint
		toAdd = append(toAdd, &c)
	}
	ret.AddAllConstraints(toAdd...)
	return ret
}

func (C *CSPTree[VAR, DOMAIN]) GetSeeds() *map[VAR]DOMAIN {
	return nil
}
