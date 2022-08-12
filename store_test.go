package guide_test

import (
	"guide"
	"testing"
)

func TestMemoryStore_GetReturnsNilOnNoGuide(t *testing.T) {
	t.Parallel()

	store := guide.OpenMemoryStore()
	coord, err := guide.NewCoordinate(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(coord)
	if err != nil {
		t.Fatal(err)
	}

	if got != nil {
		t.Error("want Store.Get to return nil")
	}
}

func TestMemoryStore_GetReturnsExistingGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate, err := guide.NewCoordinate(30, 41)
	if err != nil {
		t.Fatal(err)
	}
	want := "Tuscany"
	store.Guides[coordinate] = guide.Guide{Name: want, Coordinate: coordinate}

	g, err := store.Get(coordinate)
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
	coordinate, err := guide.NewCoordinate(30, 41)
	if err != nil {
		t.Fatal(err)
	}
	want := "Tuscany"
	g := guide.Guide{
		Name:       want,
		Coordinate: coordinate,
	}

	err = store.Create(g)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Guides[coordinate]; !ok {
		t.Error("want guide to be inserted into store")
	}
}

func TestMemoryStore_CreateExistingGuideFails(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate, err := guide.NewCoordinate(30, 41)
	if err != nil {
		t.Fatal(err)
	}
	want := "Tuscany"
	g := guide.Guide{Name: want, Coordinate: coordinate}
	store.Guides[coordinate] = g
	err = store.Create(g)

	if err == nil {
		t.Error("want Store.Create an existing habit to fail with error")
	}
}

func TestMemoryStore_UpdateHabit(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate, err := guide.NewCoordinate(30, 41)
	if err != nil {
		t.Fatal(err)
	}
	oldGuide := guide.Guide{Name: "Sicily", Coordinate: coordinate}
	store.Guides[coordinate] = oldGuide

	want := "Tuscany"
	newGuide := guide.Guide{Name: want, Coordinate: coordinate}
	err = store.Update(newGuide)
	if err != nil {
		t.Fatal(err)
	}

	if store.Guides[coordinate].Name != want {
		t.Error("want update to update guide")
	}
}

func TestMemoryStore_UpdateFailsOnNonExistingGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	coordinate, err := guide.NewCoordinate(30, 41)
	if err != nil {
		t.Fatal(err)
	}
	want := "Tuscany"
	g := guide.Guide{Name: want, Coordinate: coordinate}
	err = store.Update(g)

	if err == nil {
		t.Error("want update to fail if guide does not exist")
	}
}

func TestMemoryStore_AllGuidesReturnsSliceOfGuides(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	store.Guides = map[guide.Coordinate]guide.Guide{
		guide.Coordinate{30, 40}: guide.Guide{Name: "Tuscany"},
		guide.Coordinate{31, 41}: guide.Guide{Name: "Sicily"},
		guide.Coordinate{32, 42}: guide.Guide{Name: "Verona"},
	}

	allGuides := store.GetAllGuides()
	if len(allGuides) != len(store.Guides) {
		t.Error("want GetAllHabits to return a slice of habits")
	}
}
