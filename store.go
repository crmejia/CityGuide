package guide

import "errors"

type memoryStore struct {
	Guides  map[int]Guide
	nextKey int
}

func OpenMemoryStore() memoryStore {
	ms := memoryStore{
		Guides:  map[int]Guide{},
		nextKey: 1,
	}
	return ms
}

func (s *memoryStore) Get(id int) (*Guide, error) {
	guide, ok := s.Guides[id]
	if ok {
		return &guide, nil
	}
	return nil, nil
}

func (s *memoryStore) Create(g Guide) (int, error) {
	g.Id = s.nextKey
	s.Guides[g.Id] = g
	s.nextKey++
	return g.Id, nil
}

func (s *memoryStore) Update(g Guide) error {
	if g.Id == 0 {
		return errors.New("must set the id of the guide")
	}
	if _, ok := s.Guides[g.Id]; !ok {
		return errors.New("cannot update guide does not exist")
	}
	s.Guides[g.Id] = g
	return nil
}

// GetAllGuides returns a []Guide of all the stored guides
func (s memoryStore) GetAllGuides() []Guide {
	allGuides := make([]Guide, 0, len(s.Guides))
	for _, h := range s.Guides {
		allGuides = append(allGuides, h)
	}
	return allGuides
}
