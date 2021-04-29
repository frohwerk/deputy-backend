package predicates

func Not(p Predicate) Predicate {
	return func(path string) bool { return !p(path) }
}
