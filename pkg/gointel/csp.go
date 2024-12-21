package gointel

import (
	"sort"
)

const CSP_MAX_THREAD_COUNT = 200

func getSortedVariables[VAR comparable, DOMAIN comparable](variables []VAR, constraints map[VAR][]LocalConstraint[VAR, DOMAIN]) []VAR {
	if len(constraints) == 0 {
		return variables
	}
	cloned := make([]VAR, len(variables))
	copy(cloned, variables)
	sort.Slice(cloned, func(i, j int) bool {
		return len(constraints[variables[i]]) < len(constraints[variables[j]])
	})
	return cloned
}

type VarDomainMapCollection[VAR comparable, DOMAIN comparable] interface {
	DomainMap() map[VAR][]DOMAIN
	Variables() []VAR
	Domains() []DOMAIN // In same order as variables
	Contains(VAR) bool
}

type CSP[VAR comparable, DOMAIN comparable] interface {
	DomainMap() map[VAR][]DOMAIN
	SetDomainMap(map[VAR][]DOMAIN)
	Preprocess()
	Variables() []VAR
	GetLocalConstraints() map[VAR][]LocalConstraint[VAR, DOMAIN]
	FindAllSolutions() []map[VAR]DOMAIN
	FindOneSolution() map[VAR]DOMAIN
	IsReusableConsistent() bool
}
