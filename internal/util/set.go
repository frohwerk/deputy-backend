package util

type Set map[string]interface{}

func (s *Set) Put(v string) {
	(*s)[v] = nil
}

func (s *Set) Clear() {
	*s = make(Set)
}

func (s *Set) Contains(v string) bool {
	_, found := (*s)[v]
	return found
}

func (s *Set) Slice() []string {
	i := 0
	keys := make([]string, len(*s))
	for key := range *s {
		keys[i] = key
		i++
	}
	return keys
}
