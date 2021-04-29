package predicates

type Predicate func(path string) bool

type Predicates []Predicate

func (p Predicates) Applies(path string) bool {
	for _, f := range p {
		if f(path) {
			return true
		}
	}
	return false
}
