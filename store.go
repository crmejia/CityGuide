package guide

import "errors"

type MemoryStore struct {
	Guides map[Coordinate]Guide
}

func OpenMemoryStore() MemoryStore {
	ms := MemoryStore{Guides: map[Coordinate]Guide{}}
	return ms
}

func (s *MemoryStore) Get(coord Coordinate) (*Guide, error) {
	guide, ok := s.Guides[coord]
	if ok {
		return &guide, nil
	}
	return nil, nil
}

func (s *MemoryStore) Create(g Guide) error {
	if _, ok := s.Guides[g.Coordinate]; ok {
		return errors.New("guide already exists")
	}
	s.Guides[g.Coordinate] = g
	return nil
}

func (s *MemoryStore) Update(g Guide) error {
	if _, ok := s.Guides[g.Coordinate]; !ok {
		return errors.New("cannot update guide does not exist")
	}
	s.Guides[g.Coordinate] = g
	return nil
}

// GetAllGuides returns a []Guide of all the stored guides
func (s MemoryStore) GetAllGuides() []Guide {
	allGuides := make([]Guide, 0, len(s.Guides))
	for _, h := range s.Guides {
		allGuides = append(allGuides, h)
	}
	return allGuides
}
