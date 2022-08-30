package guide_test

import (
	"guide"
	"testing"
)

func TestMemoryStore_GetReturnsErrorOnNoGuide(t *testing.T) {
	t.Parallel()

	store := guide.OpenMemoryStore()
	_, err := store.Get(1)
	if err == nil {
		t.Errorf("want error on no guide")
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
	id := int64(44)
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
	store.Guides = map[int64]guide.Guide{
		34:    guide.Guide{Name: "Tuscany"},
		88888: guide.Guide{Name: "Sicily"},
		22:    guide.Guide{Name: "Verona"},
	}

	allGuides := store.GetAllGuides()
	if len(allGuides) != len(store.Guides) {
		t.Error("want GetAllHabits to return a slice of habits")
	}
}

func TestOpenDBStoreErrorsOnEmptyDBSource(t *testing.T) {
	t.Parallel()
	_, err := guide.OpenSQLiteStore("")
	if err == nil {
		t.Error("Want error on empty string db source")
	}
}

func TestDBStore_RoundtripCreateUpdateGet(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "rountrip.db"
	db, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := db.Create(g)
	if err != nil {
		t.Fatal(err)
	}
	want := "testGuide"
	g.Name = want
	g.Id = gid
	err = db.UpdateGuide(g)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.Get(gid)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
}

//
//func TestDBStore_AllHabits(t *testing.T) {
//	t.Parallel()
//	dbSource := t.TempDir() + "test.db"
//	sqliteStore, err := habit.OpenSQLiteStore(dbSource)
//	if err != nil {
//		t.Fatal(err)
//	}
//	habits := []*habit.Habit{
//		{Name: "piano"},
//		{Name: "surfing"},
//	}
//
//	for _, h := range habits {
//		err := sqliteStore.Create(h)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	got := sqliteStore.GetAllHabits()
//	if len(got) != len(habits) {
//		t.Errorf("want GetAllHabits to return %d habits, got %d", len(habits), len(got))
//	}
