package gointel

import (
	"context"
	"github.com/mtresnik/goutils/pkg/goutils"
	"sync"
)

type CSPDomain[VAR comparable, DOMAIN comparable] struct {
	DomainMap         map[VAR][]DOMAIN
	Preprocessors     []CSPPreprocessor[VAR, DOMAIN]
	localConstraints  map[VAR][]LocalConstraint[VAR, DOMAIN]
	globalConstraints []GlobalConstraint[VAR, DOMAIN]
}

func NewCSPDomain[VAR comparable, DOMAIN comparable](domain map[VAR][]DOMAIN, preprocessors ...CSPPreprocessor[VAR, DOMAIN]) *CSPDomain[VAR, DOMAIN] {
	return &CSPDomain[VAR, DOMAIN]{
		DomainMap:         domain,
		Preprocessors:     preprocessors,
		localConstraints:  map[VAR][]LocalConstraint[VAR, DOMAIN]{},
		globalConstraints: []GlobalConstraint[VAR, DOMAIN]{},
	}
}

func (C *CSPDomain[VAR, DOMAIN]) GetDomainMap() map[VAR][]DOMAIN {
	return C.DomainMap
}

func (C *CSPDomain[VAR, DOMAIN]) SetDomainMap(m map[VAR][]DOMAIN) {
	C.DomainMap = m
}

func (C *CSPDomain[VAR, DOMAIN]) GetVariables() []VAR {
	return goutils.Keys(C.DomainMap)
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
}

func (C *CSPDomain[VAR, DOMAIN]) GetLocalConstraints() map[VAR][]LocalConstraint[VAR, DOMAIN] {
	return C.localConstraints
}

func (C *CSPDomain[VAR, DOMAIN]) GetGlobalConstraints() []GlobalConstraint[VAR, DOMAIN] {
	return C.globalConstraints
}

func (C *CSPDomain[VAR, DOMAIN]) AddConstraint(constraint Constraint[VAR, DOMAIN]) {
	local := constraint.AsLocal()
	if local != nil {
		for _, variable := range (*local).GetVariables() {
			variableIndex := goutils.IndexOf(C.GetVariables(), func(v VAR) bool {
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

func (C *CSPDomain[VAR, DOMAIN]) AddAllConstraints(constraints ...Constraint[VAR, DOMAIN]) {
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

	variables := C.GetVariables()
	if len(variables) == 0 {
		return nil
	}
	first := variables[0]
	_, ok := C.DomainMap[first]
	if !ok {
		return nil
	}

	result := make(chan cspDomainResult[VAR, DOMAIN])

	var wg sync.WaitGroup

	for _, agent := range agents {
		wg.Add(1)
		go func(agent *CSPAgent[VAR, DOMAIN]) {
			defer wg.Done()
			solutionChannel := agent.GenerateSolutionChannel()
			select {
			case <-ctx.Done():
				return
			case result <- cspDomainResult[VAR, DOMAIN]{
				agent:    agent,
				solution: <-solutionChannel,
			}:
			}
		}(&agent)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	loop := true
	for loop {
		select {
		case task := <-result:
			if task.err != nil {
				continue
			}
			cancel()
			return task.solution

		case <-done:
			loop = false
			break
		}
	}

	<-done
	return nil
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
	_, ok := C.DomainMap[first]
	if !ok {
		return []CSPAgent[VAR, DOMAIN]{}
	}
	firstDomain, ok := C.DomainMap[first]
	if !ok {
		return []CSPAgent[VAR, DOMAIN]{}
	}

	retSlice := []CSPAgent[VAR, DOMAIN]{}

	for _, domain := range firstDomain {
		current := NewCSPNode(first, domain)
		agent := NewCSPAgent(C.DomainMap, variables, []CSPNode[VAR, DOMAIN]{*current})
		for _, local := range C.localConstraints {
			toAdd := []Constraint[VAR, DOMAIN]{}
			for _, constraint := range local {
				toAdd = append(toAdd, constraint)
			}
			agent.AddAllConstraints(toAdd...)
		}
		toAdd := []Constraint[VAR, DOMAIN]{}
		for _, constraint := range C.globalConstraints {
			toAdd = append(toAdd, constraint)
		}
		agent.AddAllConstraints(toAdd...)
		retSlice = append(retSlice, *agent)
	}

	return retSlice
}
