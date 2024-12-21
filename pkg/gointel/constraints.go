package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
	"math"
)

type Constraint[VAR comparable, DOMAIN comparable] interface {
	IsSatisfied(map[VAR]DOMAIN) bool
	IsLocal() bool
	IsReusable() bool
}

type HeuristicConstraint[VAR comparable, DOMAIN comparable] interface {
	Evaluate(map[VAR]DOMAIN) float64
	Constraint[VAR, DOMAIN]
}

type LocalConstraint[VAR comparable, DOMAIN comparable] interface {
	IsPossiblySatisfied(map[VAR]DOMAIN) bool
	GetVariables() []VAR
	Constraint[VAR, DOMAIN]
}

type LocalHeuristicConstraint[VAR comparable, DOMAIN comparable] interface {
	Evaluate(map[VAR]DOMAIN) float64
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

type GlobalConstraint[VAR comparable, DOMAIN comparable] interface {
	Constraint[VAR, DOMAIN]
}

type GlobalAllDifferentConstraint[VAR comparable, DOMAIN comparable] struct {
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	values := []DOMAIN{}
	keys := []VAR{}
	for key, value := range assignment {
		valueIndex := goutils.IndexOf(values, func(domain DOMAIN) bool {
			return domain == value
		})
		if valueIndex == -1 {
			values = append(values, value)
		}
		keyIndex := goutils.IndexOf(keys, func(key2 VAR) bool {
			return key2 == key
		})
		if keyIndex != -1 {
			keys = append(keys, key)
		}
	}
	return len(keys) == len(values)
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) IsLocal() bool {
	return false
}

func (g *GlobalAllDifferentConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
}

type LocalAllDifferentConstraint[VAR comparable, DOMAIN comparable] struct {
	Variables []VAR
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsPossiblySatisfied(assignment map[VAR]DOMAIN) bool {
	values := []DOMAIN{}
	keys := []VAR{}
	for key, value := range assignment {
		valueIndex := goutils.IndexOf(values, func(domain DOMAIN) bool {
			return domain == value
		})
		if valueIndex == -1 {
			values = append(values, value)
		}
		keyIndex := goutils.IndexOf(keys, func(key2 VAR) bool {
			return key2 == key
		})
		if keyIndex != -1 {
			keys = append(keys, key)
		}
	}
	return len(keys) == len(values)
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) GetVariables() []VAR {
	return l.Variables
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsLocal() bool {
	return true
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsSatisfied(assignment map[VAR]DOMAIN) bool {
	return l.IsPossiblySatisfied(assignment)
}

func (l *LocalAllDifferentConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
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

func (c *CardinalityConstraint[VAR, DOMAIN]) IsLocal() bool {
	return true
}

func (c *CardinalityConstraint[VAR, DOMAIN]) IsReusable() bool {
	return false
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

func (*MinimumHeuristicConstraint[VAR, DOMAIN]) IsLocal() bool {
	return true
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

// </editor-fold>
