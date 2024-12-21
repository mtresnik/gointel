package gointel

type CSPNode[VAR comparable, DOMAIN comparable] struct {
	Variable VAR
	Domain   DOMAIN
	Parent   *CSPNode[VAR, DOMAIN]
	Map      map[VAR]DOMAIN
}

func NewCSPNode[VAR comparable, DOMAIN comparable](variable VAR, domain DOMAIN, optionalParent ...*CSPNode[VAR, DOMAIN]) *CSPNode[VAR, DOMAIN] {
	var parent *CSPNode[VAR, DOMAIN] = nil
	if len(optionalParent) > 0 {
		parent = optionalParent[0]
	}
	Map := map[VAR]DOMAIN{}
	if parent != nil {
		for key, value := range parent.Map {
			Map[key] = value
		}
		Map[variable] = domain
	}
	Map[variable] = domain
	return &CSPNode[VAR, DOMAIN]{
		Variable: variable,
		Domain:   domain,
		Parent:   parent,
		Map:      Map,
	}

}
