package gointel

type CSP[VAR comparable, DOMAIN comparable] interface {
	DomainMap() map[VAR]DOMAIN
	Preprocess()
	Variables() []VAR
}
