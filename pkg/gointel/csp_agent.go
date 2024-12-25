package gointel

import (
	"container/heap"
	"context"
	"fmt"
	"github.com/mtresnik/goutils/pkg/goutils"
	"sort"
	"sync"
)

type CSPAgent[VAR comparable, DOMAIN comparable] struct {
	SortedVariables   []VAR
	DomainMap         *map[VAR][]DOMAIN
	Stack             []CSPNode[VAR, DOMAIN]
	localConstraints  map[VAR][]*LocalConstraint[VAR, DOMAIN]
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]
	SortingFunction   *func(a, b VAR) bool
}

func NewCSPAgent[VAR comparable, DOMAIN comparable](domainMap *map[VAR][]DOMAIN, sortedVariables []VAR, stack []CSPNode[VAR, DOMAIN]) *CSPAgent[VAR, DOMAIN] {
	return &CSPAgent[VAR, DOMAIN]{
		SortedVariables:   sortedVariables,
		DomainMap:         domainMap,
		Stack:             stack,
		localConstraints:  map[VAR][]*LocalConstraint[VAR, DOMAIN]{},
		globalConstraints: []*GlobalConstraint[VAR, DOMAIN]{},
	}
}

func (C *CSPAgent[VAR, DOMAIN]) AddConstraint(constraint *Constraint[VAR, DOMAIN]) {
	if constraint == nil {
		return
	}
	local := (*constraint).AsLocal()
	if local != nil {
		for _, variable := range (*local).GetVariables() {
			variableIndex := goutils.IndexOf(C.SortedVariables, func(v VAR) bool {
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
		var global GlobalConstraint[VAR, DOMAIN] = *constraint
		C.globalConstraints = append(C.globalConstraints, &global)
	}
}

func (C *CSPAgent[VAR, DOMAIN]) AddAllConstraints(constraints ...*Constraint[VAR, DOMAIN]) {
	for _, constraint := range constraints {
		C.AddConstraint(constraint)
	}
}

func GetLegalValues[VAR comparable, DOMAIN comparable](
	variable VAR,
	currentMap map[VAR]DOMAIN,
	domain []DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN],
) []DOMAIN {
	legalValues := []DOMAIN{}

	for _, value := range domain {
		isLegal := true

		if constraints, exists := localConstraints[variable]; exists {
			for _, constraint := range constraints {
				currentMap[variable] = value
				if !(*constraint).IsSatisfied(currentMap) {
					delete(currentMap, variable)
					isLegal = false
					break
				}
				delete(currentMap, variable)
			}
		}

		if isLegal {
			for _, constraint := range globalConstraints {
				currentMap[variable] = value
				if !(*constraint).IsSatisfied(currentMap) {
					delete(currentMap, variable)
					isLegal = false
					break
				}
				delete(currentMap, variable)
			}
		}

		// If legal, add the value to the resulting list
		if isLegal {
			legalValues = append(legalValues, value)
		}
	}

	return legalValues
}

// CSPNodeHeap is the min-heap for CSPNodes
type CSPNodeHeap[VAR comparable, DOMAIN comparable] []CSPNode[VAR, DOMAIN]

// Implement heap.Interface (Len, Less, Swap, Push, Pop)
func (h CSPNodeHeap[VAR, DOMAIN]) Len() int { return len(h) }
func (h CSPNodeHeap[VAR, DOMAIN]) Less(i, j int) bool {
	// Custom comparison logic: Min-Heap based on remaining legal values
	remainingValuesI := len(h[i].LegalValues) // Assume LegalValues is pre-computed/legal values cached
	remainingValuesJ := len(h[j].LegalValues) // Customize according to `GetLegalValues`
	return remainingValuesI < remainingValuesJ
}
func (h CSPNodeHeap[VAR, DOMAIN]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push adds an element to the heap
func (h *CSPNodeHeap[VAR, DOMAIN]) Push(x interface{}) {
	*h = append(*h, x.(CSPNode[VAR, DOMAIN]))
}

// Pop removes and returns the element with the smallest priority
func (h *CSPNodeHeap[VAR, DOMAIN]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (C *CSPAgent[VAR, DOMAIN]) generateStateKey(assignment map[VAR]DOMAIN) string {
	// Create a sorted representation of the map keys
	keys := make([]VAR, 0, len(assignment))
	for k := range assignment {
		keys = append(keys, k)
	}
	// Sort keys to ensure deterministic output
	sort.Slice(keys, func(i, j int) bool {
		return (*C.SortingFunction)(keys[i], keys[j])
	})

	// Build a string representation of the key-value pairs
	state := ""
	for _, k := range keys {
		state += fmt.Sprintf("%v=%v;", k, assignment[k])
	}
	return state
}

func (C *CSPAgent[VAR, DOMAIN]) GenerateSolutionChannel() chan map[VAR]DOMAIN {
	ch := make(chan map[VAR]DOMAIN)

	go func() {
		defer close(ch)

		// Create and initialize the heap
		nodeHeap := &CSPNodeHeap[VAR, DOMAIN]{}
		heap.Init(nodeHeap)

		// Add initial stack nodes to the heap
		for _, node := range C.Stack {
			heap.Push(nodeHeap, node)
		}

		var current CSPNode[VAR, DOMAIN]

		for nodeHeap.Len() > 0 {
			// Pop the node with the least priority (minimum remaining values)
			current = heap.Pop(nodeHeap).(CSPNode[VAR, DOMAIN])
			currentMap := current.GetMap()

			// Check local consistency
			if IsLocallyConsistent(current.Variable, currentMap, C.GetLocalConstraints(), C.GetGlobalConstraints()) {
				// If all variables are assigned, check global consistency and yield solution
				if len(currentMap) == len(C.GetVariables()) {
					if IsConsistent(current.Variable, currentMap, C.GetLocalConstraints(), C.GetGlobalConstraints()) {
						ch <- currentMap
					}
					continue
				} else {
					// Get the next variable to process
					nextVariableIndex := goutils.IndexOf(C.GetVariables(), func(v VAR) bool {
						_, assigned := current.GetMap()[v]
						return !assigned
					})
					if nextVariableIndex == -1 {
						continue
					}

					nextVariable := C.GetVariables()[nextVariableIndex]
					nextDomain, ok := C.GetDomainMap()[nextVariable]
					if !ok {
						continue
					}

					// Reduce the domain and add subproblems to the heap
					reduced := ReduceDomain(nextVariable, currentMap, nextDomain, C.GetLocalConstraints(), C.GetGlobalConstraints())
					for _, domain := range reduced {
						heap.Push(nodeHeap, *NewCSPNode(nextVariable, nextVariableIndex, domain, GetLegalValues(nextVariable, currentMap, C.GetDomainForVariable(nextVariable), C.localConstraints, C.globalConstraints), &current))
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

func (C *CSPAgent[VAR, DOMAIN]) GetLocalConstraints() map[VAR][]*LocalConstraint[VAR, DOMAIN] {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	solutionChannel := C.GenerateSolutionChannel()

	for {
		select {
		case <-ctx.Done(): // Stop processing if the context is canceled
			return nil
		case solution, ok := <-solutionChannel:
			if !ok {
				// Channel closed, no more solutions
				return nil
			}
			if solution != nil {
				// Cancel any ongoing solution generation
				cancel()
				return solution
			}
		}
	}
}

func FindFirstSolutionFromAgents[VAR comparable, DOMAIN comparable](ctx context.Context, agents []CSPAgent[VAR, DOMAIN]) map[VAR]DOMAIN {
	results := make(chan map[VAR]DOMAIN)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	// Start agents to generate solutions in parallel
	for _, agent := range agents {
		wg.Add(1)
		go func(a CSPAgent[VAR, DOMAIN]) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Stop processing when context is canceled
					return
				case solution, ok := <-a.GenerateSolutionChannel():
					if !ok {
						return
					}
					if solution != nil {
						select {
						case results <- solution:
							return
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(agent)
	}

	// Wait for all agents to finish if no solution is found
	go func() {
		wg.Wait()
		close(results) // Close results channel when all goroutines complete
	}()

	// Retrieve the first solution
	for solution := range results {
		if solution != nil {
			cancel() // Cancel all other agent processing
			return solution
		}
	}

	return nil
}

func (C *CSPAgent[VAR, DOMAIN]) GetDomainMap() map[VAR][]DOMAIN {
	return *C.DomainMap
}

func (C *CSPAgent[VAR, DOMAIN]) SetDomainMap(m map[VAR][]DOMAIN) {
	C.DomainMap = &m
}

func (C *CSPAgent[VAR, DOMAIN]) GetVariables() []VAR {
	return C.SortedVariables
}

func (C *CSPAgent[VAR, DOMAIN]) GetDomainForVariable(variable VAR) []DOMAIN {
	return (*C.DomainMap)[variable]
}

func (C *CSPAgent[VAR, DOMAIN]) Contains(v VAR) bool {
	_, ok := (*C.DomainMap)[v]
	return ok
}

func (C *CSPAgent[VAR, DOMAIN]) GetGlobalConstraints() []*GlobalConstraint[VAR, DOMAIN] {
	return C.globalConstraints
}

func (C *CSPAgent[VAR, DOMAIN]) GetSeeds() *map[VAR]DOMAIN {
	return nil
}
