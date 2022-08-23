package guide_test

import (
	"guide"
	"testing"
)

func TestMemoryStore_GetReturnsNilAndErrorOnNoGuide(t *testing.T) {
	t.Parallel()

	store := guide.OpenMemoryStore()
	got, err := store.Get(1)
	if err == nil {
		t.Errorf("want error on no guide")
	}

	if got != nil {
		t.Error("want Store.Get to return nil")
	}
}

func TestMemoryStore_GetReturnsExistingGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate := guide.Coordinate{30, 40}
	want := "Tuscany"
	store.Guides[1] = guide.Guide{Name: want, Coordinate: coordinate}

	g, err := store.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		t.Fatal(err)
	}

	got := g.Name
	if want != got {
		t.Errorf("want Guide name to be %s, got %s", want, got)
	}
}

func TestMemoryStore_CreateNewGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	want := "Tuscany"
	g := guide.Guide{
		Name:       want,
		Coordinate: guide.Coordinate{10, 10},
	}

	id, err := store.Create(g)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Guides[id]; !ok {
		t.Error("want guide to be inserted into store")
	}
}

func TestMemoryStore_UpdateGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate := guide.Coordinate{30, 40}
	oldGuide := guide.Guide{Name: "Sicily", Coordinate: coordinate}
	id := 44
	store.Guides[id] = oldGuide

	want := "Tuscany"
	newGuide := guide.Guide{Id: id, Name: want, Coordinate: coordinate}
	err := store.Update(newGuide)
	if err != nil {
		t.Fatal(err)
	}

	if store.Guides[id].Name != want {
		t.Error("want update to update guide")
	}
}

func TestMemoryStore_UpdateFailsOnNonExistingGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate := guide.Coordinate{30, 40}
	want := "Tuscany"
	g := guide.Guide{Id: 12, Name: want, Coordinate: coordinate}
	err := store.Update(g)

	if err == nil {
		t.Error("want update to fail if guide does not exist")
	}
}

func TestMemoryStore_UpdateFailsOnNonSetID(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate := guide.Coordinate{30, 40}
	want := "Tuscany"
	g := guide.Guide{Id: 0, Name: want, Coordinate: coordinate}
	err := store.Update(g)

	if err == nil {
		t.Error("want update to fail if guide does not exist")
	}
}

func TestMemoryStore_AllGuidesReturnsSliceOfGuides(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	store.Guides = map[int]guide.Guide{
		34:    guide.Guide{Name: "Tuscany"},
		88888: guide.Guide{Name: "Sicily"},
		22:    guide.Guide{Name: "Verona"},
	}

	allGuides := store.GetAllGuides()
	if len(allGuides) != len(store.Guides) {
		t.Error("want GetAllHabits to return a slice of habits")
	}
}
