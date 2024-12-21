package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
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
