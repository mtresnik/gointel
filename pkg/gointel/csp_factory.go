package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
	"math"
)

const CSP_REPEAT_THRESHOLD = 3.0287511e+14

type CSPFactoryRequest[VAR comparable, DOMAIN comparable] struct {
	DomainMap     map[VAR][]DOMAIN
	MaxTime       int64
	Preprocessors []CSPPreprocessor[VAR, DOMAIN]
	IsRepeatable  bool
}

type CSPFactory[VAR comparable, DOMAIN comparable] func(request CSPFactoryRequest[VAR, DOMAIN]) *CSP[VAR, DOMAIN]

func DefaultCSPFactory[VAR comparable, DOMAIN comparable](request CSPFactoryRequest[VAR, DOMAIN]) *CSP[VAR, DOMAIN] {
	numVariables := len(request.DomainMap)
	domainValues := [][]DOMAIN{}
	for _, domain := range request.DomainMap {
		domainValues = append(domainValues, domain)
	}
	maxDomainInt := len(goutils.MaxBy(domainValues, func(domains []DOMAIN) float64 {
		return float64(len(domains))
	}))
	maxDomain := float64(maxDomainInt)
	maxChildren := math.Pow(maxDomain, float64(numVariables))
	if maxChildren > float64(CSP_MAX_THREAD_COUNT) {
		if request.IsRepeatable || maxChildren > CSP_REPEAT_THRESHOLD {
			// CSPGoRoutine
			panic("implement me")
		} else {
			// CSPInferredAsync
			panic("implement me")
		}
	}
	if maxDomainInt > numVariables {
		var ret CSP[VAR, DOMAIN] = NewCSPDomain(request.DomainMap)
		return &ret
	}
	var ret CSP[VAR, DOMAIN] = NewCSPTree(request.DomainMap)
	return &ret
}
