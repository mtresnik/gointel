package gointel

import (
	"context"
	"github.com/mtresnik/goutils/pkg/goutils"
	"sort"
	"sync"
)

type CSPDomain[VAR comparable, DOMAIN comparable] struct {
	DomainMap         map[VAR][]DOMAIN
	Preprocessors     []CSPPreprocessor[VAR, DOMAIN]
	localConstraints  map[VAR][]*LocalConstraint[VAR, DOMAIN]
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]
	Seeds             *map[VAR]DOMAIN
	sortedVariables   *[]VAR
	sortingFunction   *func(a, b VAR) bool
}

func NewCSPDomain[VAR comparable, DOMAIN comparable](domain map[VAR][]DOMAIN, preprocessors ...CSPPreprocessor[VAR, DOMAIN]) *CSPDomain[VAR, DOMAIN] {
	return &CSPDomain[VAR, DOMAIN]{
		DomainMap:         domain,
		Preprocessors:     preprocessors,
		localConstraints:  map[VAR][]*LocalConstraint[VAR, DOMAIN]{},
		globalConstraints: []*GlobalConstraint[VAR, DOMAIN]{},
		Seeds:             &map[VAR]DOMAIN{},
		sortedVariables:   nil,
	}
}

func (C *CSPDomain[VAR, DOMAIN]) GetDomainMap() map[VAR][]DOMAIN {
	return C.DomainMap
}

func (C *CSPDomain[VAR, DOMAIN]) SetDomainMap(m map[VAR][]DOMAIN) {
	C.DomainMap = m
}

func (C *CSPDomain[VAR, DOMAIN]) SetSortingFunction(function func(a, b VAR) bool) {
	C.sortingFunction = &function
}

func (C *CSPDomain[VAR, DOMAIN]) GetVariables() []VAR {
	if C.sortedVariables != nil {
		return *C.sortedVariables
	}
	retVariables := goutils.Keys(C.DomainMap)
	if C.sortingFunction != nil {
		sort.Slice(retVariables, func(i, j int) bool {
			return (*C.sortingFunction)(retVariables[i], retVariables[j])
		})
	}
	C.sortedVariables = &retVariables
	return *C.sortedVariables
}

func (C *CSPDomain[VAR, DOMAIN]) GetDomainForVariable(variable VAR) []DOMAIN {
	ret, ok := C.DomainMap[variable]
	if !ok {
		return []DOMAIN{}
	}
	return ret
}

func (C *CSPDomain[VAR, DOMAIN]) Contains(v VAR) bool {
	_, ok := C.DomainMap[v]
	return ok
}

func (C *CSPDomain[VAR, DOMAIN]) Preprocess() {
	var csp CSP[VAR, DOMAIN] = C
	for _, preprocessor := range C.Preprocessors {
		preprocessor.Preprocess(&csp)
	}
	// Reduce domains
	assignment := map[VAR]DOMAIN{}
	for variable, domains := range C.DomainMap {
		if len(domains) == 1 {
			assignment[variable] = domains[0]
			continue
		}
	}
	for variable, domains := range C.DomainMap {
		reduced := ReduceDomain(variable, assignment, domains, C.GetLocalConstraints(), C.GetGlobalConstraints())
		C.DomainMap[variable] = reduced
	}
}

func (C *CSPDomain[VAR, DOMAIN]) GetLocalConstraints() map[VAR][]*LocalConstraint[VAR, DOMAIN] {
	return C.localConstraints
}

func (C *CSPDomain[VAR, DOMAIN]) GetGlobalConstraints() []*GlobalConstraint[VAR, DOMAIN] {
	return C.globalConstraints
}

func (C *CSPDomain[VAR, DOMAIN]) AddConstraint(constraint *Constraint[VAR, DOMAIN]) {
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
				var constraintLookup []*LocalConstraint[VAR, DOMAIN] = nil
				temp, ok := C.localConstraints[variable]
				if ok {
					constraintLookup = temp
				} else {
					constraintLookup = []*LocalConstraint[VAR, DOMAIN]{}
				}
				constraintLookup = append(constraintLookup, local)
				C.localConstraints[variable] = constraintLookup
			}
		}
	} else {
		var global GlobalConstraint[VAR, DOMAIN] = *constraint
		C.globalConstraints = append(C.globalConstraints, &global)
	}
}

func (C *CSPDomain[VAR, DOMAIN]) AddAllConstraints(constraints ...*Constraint[VAR, DOMAIN]) {
	for _, constraint := range constraints {
		C.AddConstraint(constraint)
	}
}

func (C *CSPDomain[VAR, DOMAIN]) FindAllSolutions() []map[VAR]DOMAIN {
	C.Preprocess()
	agents := C.ConstructAgents()
	syncList := goutils.NewSyncList[map[VAR]DOMAIN]()

	variables := C.GetVariables()
	if len(variables) == 0 {
		return []map[VAR]DOMAIN{}
	}
	first := variables[0]
	_, ok := C.DomainMap[first]
	if !ok {
		return []map[VAR]DOMAIN{}
	}

	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go func(agent *CSPAgent[VAR, DOMAIN]) {
			defer wg.Done()
			channel := agent.GenerateSolutionChannel()
			for solution := range channel {
				syncList.Add(solution)
			}
		}(&agent)
	}
	wg.Wait()

	return syncList.ToSlice()
}

type cspDomainResult[VAR comparable, DOMAIN comparable] struct {
	solution map[VAR]DOMAIN
	agent    *CSPAgent[VAR, DOMAIN]
	err      error
}

func (C *CSPDomain[VAR, DOMAIN]) FindOneSolution() map[VAR]DOMAIN {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	C.Preprocess()
	agents := C.ConstructAgents()
	return FindFirstSolutionFromAgents(ctx, agents)
}

func (C *CSPDomain[VAR, DOMAIN]) GenerateSolutionChannel() chan map[VAR]DOMAIN {
	panic("illegal call")
}

func (C *CSPDomain[VAR, DOMAIN]) ConstructAgents() []CSPAgent[VAR, DOMAIN] {
	variables := C.GetVariables()
	if len(variables) == 0 {
		return []CSPAgent[VAR, DOMAIN]{}
	}
	first := variables[0]
	firstDomain, ok := C.DomainMap[first]
	if !ok {
		return []CSPAgent[VAR, DOMAIN]{}
	}
	if len(firstDomain) == 1 {
		first = goutils.MaxBy(variables, func(v VAR) float64 {
			currDomain, ok := C.DomainMap[v]
			if ok {
				return float64(len(currDomain))
			}
			return 0.0
		})
		firstDomain, ok = C.DomainMap[first]
		if !ok {
			return []CSPAgent[VAR, DOMAIN]{}
		}
	}
	println("Constructing agents for domain:", len(firstDomain))

	retSlice := []CSPAgent[VAR, DOMAIN]{}

	for _, domain := range firstDomain {
		current := NewCSPNode(first, 0, domain, firstDomain)
		agent := NewCSPAgent(&C.DomainMap, variables, []CSPNode[VAR, DOMAIN]{*current})
		agent.SortingFunction = C.sortingFunction
		for _, local := range C.localConstraints {
			toAdd := []*Constraint[VAR, DOMAIN]{}
			for _, constraint := range local {
				var c Constraint[VAR, DOMAIN] = *constraint
				toAdd = append(toAdd, &c)
			}
			agent.AddAllConstraints(toAdd...)
		}
		toAdd := []*Constraint[VAR, DOMAIN]{}
		for _, constraint := range C.globalConstraints {
			var c Constraint[VAR, DOMAIN] = *constraint
			toAdd = append(toAdd, &c)
		}
		agent.AddAllConstraints(toAdd...)
		retSlice = append(retSlice, *agent)
	}

	return retSlice
}

func (C *CSPDomain[VAR, DOMAIN]) GetSeeds() *map[VAR]DOMAIN {
	return C.Seeds
}
