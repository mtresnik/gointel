package gointel

import (
	"sort"
)

const CSP_MAX_CHILDREN = 200

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
	GetDomainMap() map[VAR][]DOMAIN
	SetDomainMap(map[VAR][]DOMAIN)
	GetVariables() []VAR
	GetDomainForVariable(variable VAR) []DOMAIN
	Contains(VAR) bool
}

type CSP[VAR comparable, DOMAIN comparable] interface {
	VarDomainMapCollection[VAR, DOMAIN]
	Preprocess()
	GetLocalConstraints() map[VAR][]*LocalConstraint[VAR, DOMAIN]
	GetGlobalConstraints() []*GlobalConstraint[VAR, DOMAIN]
	AddConstraint(constraint *Constraint[VAR, DOMAIN])
	AddAllConstraints(constraints ...*Constraint[VAR, DOMAIN])
	FindAllSolutions() []map[VAR]DOMAIN
	FindOneSolution() map[VAR]DOMAIN
	GetSeeds() *map[VAR]DOMAIN
}

type MultiCSP[VAR comparable, DOMAIN comparable] interface {
	CSP[VAR, DOMAIN]
	GenerateSolutionChannel() chan map[VAR]DOMAIN
}

func IsLocallyConsistent[VAR comparable, DOMAIN comparable](
	variable VAR,
	assignment map[VAR]DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]) bool {
	if len(localConstraints) == 0 {
		return len(globalConstraints) != 0
	}
	tempConstraints, ok := localConstraints[variable]
	if !ok {
		return false
	}
	for _, constraint := range tempConstraints {
		if !(*constraint).IsPossiblySatisfied(assignment) {
			return false
		}
	}
	return true
}

func IsGloballyConsistent[VAR comparable, DOMAIN comparable](assignment map[VAR]DOMAIN, globalConstraints []*GlobalConstraint[VAR, DOMAIN]) bool {
	for _, constraint := range globalConstraints {
		pass := (*constraint).IsSatisfied(assignment)
		if !pass {
			return false
		}
	}
	return true
}

func IsAbsolutelyConsistent[VAR comparable, DOMAIN comparable](
	variable VAR,
	assignment map[VAR]DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]) bool {
	if len(localConstraints) == 0 {
		return len(globalConstraints) != 0
	}
	tempConstraints, ok := localConstraints[variable]
	if !ok {
		return false
	}
	for _, constraint := range tempConstraints {
		if !(*constraint).IsSatisfied(assignment) {
			return false
		}
	}
	return true
}

func IsConsistent[VAR comparable, DOMAIN comparable](
	variable VAR,
	assignment map[VAR]DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN]) bool {
	return IsGloballyConsistent(assignment, globalConstraints) && IsAbsolutelyConsistent(variable, assignment, localConstraints, globalConstraints)
}

func IsReusableConsistent[VAR comparable, DOMAIN comparable](
	assignment map[VAR]DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN],
) bool {
	reusableConstraints := []*Constraint[VAR, DOMAIN]{}
	for _, constraints := range localConstraints {
		for _, constraint := range constraints {
			if (*constraint).IsReusable() {
				var c Constraint[VAR, DOMAIN] = *constraint
				reusableConstraints = append(reusableConstraints, &c)
			}
		}
	}
	for _, constraint := range globalConstraints {
		if (*constraint).IsReusable() {
			var c Constraint[VAR, DOMAIN] = *constraint
			reusableConstraints = append(reusableConstraints, &c)
		}
	}
	for _, constraint := range reusableConstraints {
		if !(*constraint).IsSatisfied(assignment) {
			return false
		}
	}
	return true
}

func ReduceDomain[VAR comparable, DOMAIN comparable](
	variable VAR,
	assignment map[VAR]DOMAIN, tempDomain []DOMAIN,
	localConstraints map[VAR][]*LocalConstraint[VAR, DOMAIN],
	globalConstraints []*GlobalConstraint[VAR, DOMAIN],
) []DOMAIN {
	currDomain := tempDomain
	for _, constraints := range localConstraints {
		for _, constraint := range constraints {
			currDomain = (*constraint).ReduceDomain(variable, assignment, currDomain)
		}
	}
	for _, constraint := range globalConstraints {
		currDomain = (*constraint).ReduceDomain(variable, assignment, currDomain)
	}
	return currDomain
}
