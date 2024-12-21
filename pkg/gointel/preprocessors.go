package gointel

import (
	"fmt"
	"github.com/mtresnik/goutils/pkg/goutils"
)

type CSPPreprocessor[VAR comparable, DOMAIN comparable] interface {
	Preprocess(csp *CSP[VAR, DOMAIN])
}

// AC3Preprocessor Refs https://en.wikipedia.org/wiki/AC-3_algorithm
type AC3Preprocessor[VAR comparable, DOMAIN comparable] struct {
}

func CloneMapWithSlices[K comparable, V any](original map[K][]V) map[K][]V {
	cloned := make(map[K][]V, len(original))
	for key, value := range original {
		newSlice := make([]V, len(value))
		copy(newSlice, value)
		cloned[key] = newSlice
	}
	return cloned
}

func CloneMap[K comparable, V any](original map[K]V) map[K]V {
	cloned := make(map[K]V, len(original))
	for key, value := range original {
		cloned[key] = value
	}
	return cloned
}

func (A *AC3Preprocessor[VAR, DOMAIN]) Preprocess(cspPtr *CSP[VAR, DOMAIN]) {
	currentDomain := map[VAR][]DOMAIN{}
	if cspPtr == nil {
		return
	}
	csp := *cspPtr
	originalDomain := CloneMapWithSlices(csp.GetDomainMap())
	originalConstraints := csp.GetLocalConstraints()
	variables := csp.Variables()

	unaryConstraints := map[VAR][]LocalConstraint[VAR, DOMAIN]{}
	binaryConstraints := map[VAR][]LocalConstraint[VAR, DOMAIN]{}
	for _, variable := range variables {
		unaryConstraints[variable] = []LocalConstraint[VAR, DOMAIN]{}
		binaryConstraints[variable] = []LocalConstraint[VAR, DOMAIN]{}
	}

	for variable, constraints := range originalConstraints {
		for _, constraint := range constraints {
			if IsUnary(constraint) {
				unaryConstraints[variable] = append(unaryConstraints[variable], constraint)
			} else if IsBinary(constraint) {
				binaryConstraints[variable] = append(binaryConstraints[variable], constraint)
			}
		}
	}

	// Find all domains that work for the unary constraints
	for _, variable := range variables {
		currentDomain[variable] = goutils.Filter(originalDomain[variable], func(domain DOMAIN) bool {
			all := true
			for _, constraint := range unaryConstraints[variable] {
				if !constraint.IsPossiblySatisfied(map[VAR]DOMAIN{variable: domain}) {
					all = false
					break
				}
			}
			return all
		})
	}

	// Find all domains that work for binary constraints
	workQueue := []LocalConstraint[VAR, DOMAIN]{}
	for _, it := range binaryConstraints {
		workQueue = append(workQueue, it...)
	}
	println("WorkQueue size:", len(workQueue))

	for {
		if len(workQueue) > 0 {
			constraint := workQueue[0]
			workQueue = workQueue[1:]
			x, y := constraint.GetVariables()[0], constraint.GetVariables()[1]
			shared := getSharedConstraints(x, y, binaryConstraints)
			if arcReduce(x, y, shared, currentDomain) && len(currentDomain) == 0 {
				panic(fmt.Sprintf("variable %s has an empty domain", x))
			}
		}
		if len(workQueue) == 0 {
			break
		}
	}

	// Set values for CSP
	newDomain := CloneMap(currentDomain)
	csp.SetDomainMap(newDomain)
}

func arcReduce[VAR comparable, DOMAIN comparable](x VAR, y VAR, binaryConstraints []LocalConstraint[VAR, DOMAIN], currentDomain map[VAR][]DOMAIN) bool {
	change := false
	currentDomainX := currentDomain[x]
	currentDomainY := currentDomain[y]
	for index := len(currentDomainX) - 1; index >= 0; index-- {
		vx := currentDomainX[index]
		vyIndex := -1
		vyIndex = goutils.IndexOf(currentDomainY, func(yDomain DOMAIN) bool {
			all := true
			for _, constraint := range binaryConstraints {
				if !constraint.IsPossiblySatisfied(map[VAR]DOMAIN{x: vx, y: yDomain}) {
					all = false
					break
				}
			}
			return all
		})
		if vyIndex == -1 {
			currentDomainX = append(currentDomainX[:index], currentDomainX[index+1:]...)
			change = true
		}
	}
	currentDomain[x] = currentDomainX
	return change
}

func getSharedConstraints[VAR comparable, DOMAIN comparable](x VAR, y VAR, allConstraints map[VAR][]LocalConstraint[VAR, DOMAIN]) []LocalConstraint[VAR, DOMAIN] {
	retSlice := []LocalConstraint[VAR, DOMAIN]{}
	sliceX := allConstraints[x]
	sliceY := allConstraints[y]
	retSlice = append(retSlice, goutils.Filter(sliceX, func(constraint LocalConstraint[VAR, DOMAIN]) bool {
		return goutils.IndexOf(constraint.GetVariables(), func(variable VAR) bool {
			return variable == y
		}) != -1
	})...)
	retSlice = append(retSlice, goutils.Filter(sliceY, func(constraint LocalConstraint[VAR, DOMAIN]) bool {
		return goutils.IndexOf(constraint.GetVariables(), func(variable VAR) bool {
			return variable == x
		}) != -1
	})...)
	return retSlice
}
