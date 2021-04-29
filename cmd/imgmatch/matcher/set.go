package matcher

type set map[string]*struct{}

func (s *set) put(v string) {
	(*s)[v] = nil
}

func (s *set) clear() {
	*s = make(set)
}

func (s *set) contains(v string) bool {
	_, found := (*s)[v]
	return found
}
