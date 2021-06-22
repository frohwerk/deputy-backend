package dependencies

type memoryStore map[string][]string

func (r *memoryStore) Direct(id string) ([]string, error) {
	if deps, ok := (*r)[id]; ok {
		return deps, nil
	}
	return []string{}, nil
}

type Cache struct {
	Store
	entries map[string][]string
}

func (s *Cache) Direct(id string) ([]string, error) {
	if s.entries == nil {
		s.entries = make(map[string][]string)
	}

	if v, ok := s.entries[id]; ok {
		return v, nil
	}

	v, err := s.Store.Direct(id)
	if err != nil {
		return nil, err
	}

	s.entries[id] = v
	return v, nil
}
