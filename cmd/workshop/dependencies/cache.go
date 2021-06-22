package dependencies

type memoryStore map[string][]string

func (r *memoryStore) Direct(id string) ([]string, error) {
	if deps, ok := (*r)[id]; ok {
		return deps, nil
	}
	return []string{}, nil
}

type cache struct {
	Store
	entries map[string][]string
}

func Cache(s Store) Store {
	return &cache{s, make(map[string][]string)}
}

func (s *cache) Direct(id string) ([]string, error) {
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
