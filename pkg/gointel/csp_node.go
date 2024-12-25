package gointel

type CSPNode[VAR comparable, DOMAIN comparable] struct {
	Variable      VAR
	VariableIndex int
	Domain        DOMAIN
	Parent        *CSPNode[VAR, DOMAIN]
	cachedMap     *map[VAR]DOMAIN
	LegalValues   []DOMAIN
}

func NewCSPNode[VAR comparable, DOMAIN comparable](variable VAR, variableIndex int, domain DOMAIN, legalValues []DOMAIN, optionalParent ...*CSPNode[VAR, DOMAIN]) *CSPNode[VAR, DOMAIN] {
	var parent *CSPNode[VAR, DOMAIN] = nil
	if len(optionalParent) > 0 {
		parent = optionalParent[0]
	}
	newMap := make(map[VAR]DOMAIN)
	if parent != nil {
		// Copy assignment from parent map
		for k, v := range parent.GetMap() {
			newMap[k] = v
		}
	}
	// Add the current variable assignment
	newMap[variable] = domain
	return &CSPNode[VAR, DOMAIN]{
		Variable:      variable,
		VariableIndex: variableIndex,
		Domain:        domain,
		Parent:        parent,
		cachedMap:     &newMap,
		LegalValues:   legalValues,
	}

}

func (node *CSPNode[VAR, DOMAIN]) GetMap() map[VAR]DOMAIN {
	return *node.cachedMap
}
