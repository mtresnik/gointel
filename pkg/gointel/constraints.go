package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
	"math"
)

type Constraint[VAR comparable, DOMAIN comparable] interface {
	IsSatisfied(assignment map[VAR]DOMAIN) bool
	AsLocal() *LocalConstraint[VAR, DOMAIN]
	IsReusable() bool
	ReduceDomain(variable VAR, assignment map[VAR]DOMAIN, domain []DOMAIN) []DOMAIN
}

type HeuristicConstraint[VAR comparable, DOMAIN comparable] interface {
	Constraint[VAR, DOMAIN]
	Evaluate(assignment map[VAR]DOMAIN) float64
}

type LocalConstraint[VAR comparable, DOMAIN comparable] interface {
	Constraint[VAR, DOMAIN]
	IsPossiblySatisfied(assignment map[VAR]DOMAIN) bool
	GetVariables() []VAR
}

type LocalHeuristicConstraint[VAR comparable, DOMAIN comparable] interface {
	Evaluate(assignment map[VAR]DOMAIN) float64
	LocalConstraint[VAR, DOMAIN]
}

func IsUnary[VAR comparable, DOMAIN comparable](constraint LocalConstraint[VAR, DOMAIN]) bool {
	return len(constraint.GetVariables()) == 1
}

func IsBinary[VAR comparable, DOMAIN comparable](constraint LocalConstraint[VAR, DOMAIN]) bool {
	return len(constraint.GetVariables()) == 2
}

func IsTernary[VAR comparable, DOMAIN comparable](constraint LocalConstraint[VAR, DOMAIN]) bool {
	return len(constraint.GetVariables()) == 3
}

type GlobalConstraint[VAR comparable, DOMAIN comparable] Constraint[VAR, DOMAIN]

type GlobalAllDifferentConstraint[VAR comparable, DOMAIN comparable] struct {
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	valueSet := map[DOMAIN]bool{}
	keySet := map[VAR]bool{}
	for key, value := range assignment {
		valueSet[value] = true
		keySet[key] = true
	}
	return len(keySet) == len(valueSet)
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) AsLocal() *LocalConstraint[VAR, DOMAIN] {
	return nil
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) ReduceDomain(variable VAR, assignment map[VAR]DOMAIN, domain []DOMAIN) []DOMAIN {
	return domain
}

type LocalAllDifferentConstraint[VAR comparable, DOMAIN comparable] struct {
	Variables *[]VAR
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsPossiblySatisfied(assignment map[VAR]DOMAIN) bool {
	variableSet := map[VAR]bool{}
	domainSet := map[DOMAIN]bool{}
	for _, variable := range *(l.Variables) {
		if domain, ok := assignment[variable]; ok {
			variableSet[variable] = true
			if goutils.SetContains(domainSet, domain) {
				return false
			}
			domainSet[domain] = true
		}
	}
	return true
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) GetVariables() []VAR {
	return *l.Variables
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) AsLocal() *LocalConstraint[VAR, DOMAIN] {
	var localConstraint LocalConstraint[VAR, DOMAIN] = l
	return &localConstraint
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	return l.IsPossiblySatisfied(assignment)
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) ReduceDomain(nextVariable VAR, assignment map[VAR]DOMAIN, domain []DOMAIN) []DOMAIN {
	if len(domain) == 0 {
		return domain
	}
	retDomain := []DOMAIN{}
	for _, currDomain := range domain {
		assignment[nextVariable] = currDomain
		if l.IsPossiblySatisfied(assignment) {
			retDomain = append(retDomain, currDomain)
		}
	}
	delete(assignment, nextVariable)
	return retDomain
}

type SumConstraint[VAR comparable, DOMAIN comparable] interface {
	Add(one, other DOMAIN) DOMAIN
	MaxValue() DOMAIN
	Constraint[VAR, DOMAIN]
}

type IntSumConstraint[VAR comparable] interface {
	SumConstraint[VAR, int]
}

type FloatSumConstraint[VAR comparable] interface {
	SumConstraint[VAR, float64]
}

// CardinalityConstraint <editor-fold>
type CardinalityConstraint[VAR comparable, DOMAIN comparable] struct {
	Variables []VAR
	MaxCount  int
	Domain    DOMAIN
}

func NewCardinalityConstraint[VAR comparable, DOMAIN comparable](variables []VAR, maxCount int, domain DOMAIN) *CardinalityConstraint[VAR, DOMAIN] {
	return &CardinalityConstraint[VAR, DOMAIN]{
		Variables: variables,
		MaxCount:  maxCount,
		Domain:    domain,
	}
}

func (c *CardinalityConstraint[VAR, DOMAIN]) IsPossiblySatisfied(assignment map[VAR]DOMAIN) bool {
	count := 0
	for _, d := range assignment {
		if d == c.Domain {
			count++
		}
		if count > c.MaxCount {
			return false
		}
	}
	return true
}

func (c *CardinalityConstraint[VAR, DOMAIN]) GetVariables() []VAR {
	return c.Variables
}

func (c *CardinalityConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	return c.IsPossiblySatisfied(assignment)
}

func (c *CardinalityConstraint[VAR, DOMAIN]) AsLocal() *LocalConstraint[VAR, DOMAIN] {
	var localConstraint LocalConstraint[VAR, DOMAIN] = c
	return &localConstraint
}

func (c *CardinalityConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
}

func (c *CardinalityConstraint[VAR, DOMAIN]) ReduceDomain(variable VAR, assignment map[VAR]DOMAIN, domain []DOMAIN) []DOMAIN {
	return domain
}

// </editor-fold>

// MinimumHeuristicConstraint <editor-fold>
type MinimumHeuristicConstraint[VAR comparable, DOMAIN comparable] struct {
	Variables []VAR
	Evaluator func(assignment map[VAR]DOMAIN) float64
	minValue  float64
}

func NewMinimumHeuristicConstraint[VAR comparable, DOMAIN comparable](variables []VAR, evaluator func(map[VAR]DOMAIN) float64) *MinimumHeuristicConstraint[VAR, DOMAIN] {
	return &MinimumHeuristicConstraint[VAR, DOMAIN]{
		Variables: variables,
		Evaluator: evaluator,
		minValue:  math.MaxFloat64,
	}
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) GetVariables() []VAR {
	return m.Variables
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) AsLocal() *LocalConstraint[VAR, DOMAIN] {
	var localConstraint LocalConstraint[VAR, DOMAIN] = m
	return &localConstraint
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) Evaluate(assignment map[VAR]DOMAIN) float64 {
	return m.Evaluator(assignment)
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) IsPossiblySatisfied(assignment map[VAR]DOMAIN) bool {
	return m.Evaluate(assignment) <= m.minValue
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	tempValue := m.Evaluator(assignment)
	if tempValue <= m.minValue {
		m.minValue = tempValue
		return true
	}
	return false
}

func (m *MinimumHeuristicConstraint[VAR, DOMAIN]) ReduceDomain(variable VAR, assignment map[VAR]DOMAIN, domain []DOMAIN) []DOMAIN {
	return domain
}

// </editor-fold>
