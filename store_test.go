package guide_test

import (
	"guide"
	"testing"
)

func TestMemoryStore_GetReturnsErrorOnNoGuide(t *testing.T) {
	t.Parallel()

	store := guide.OpenMemoryStore()
	_, err := store.GetGuide(1)
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

	g, err := store.GetGuide(1)
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

	id, err := store.CreateGuide(g)
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
	err := store.UpdateGuide(newGuide)
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
	err := store.UpdateGuide(g)

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
	err := store.UpdateGuide(g)

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

func TestMemoryStore_PoiRoundtripCreateUpdateGetPoi(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := store.CreateGuide(g)
	if err != nil {
		t.Fatal(err)
	}

	poi, err := guide.NewPointOfInterest("test", gid, guide.PoiWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := store.CreatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}
	want := "testPOI"
	poi.Id = pid
	poi.Name = want
	err = store.UpdatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetPoi(pid)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
}

func TestMemoryStore_GetAllPois(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := store.CreateGuide(g)
	if err != nil {
		t.Fatal(err)
	}
	pois := []string{"A", "B", "C"}
	for _, name := range pois {
		poi, err := guide.NewPointOfInterest(name, gid, guide.PoiWithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		store.CreatePoi(poi)
	}
	got := store.GetAllPois(gid)
	if len(got) != len(pois) {
		t.Errorf("want GetAllPois to return %d points of interest, got %d", len(pois), len(got))
	}
}

func TestOpenSQLiteStoreErrorsOnEmptyDBSource(t *testing.T) {
	t.Parallel()
	_, err := guide.OpenSQLiteStore("")
	if err == nil {
		t.Error("Want error on empty string store source")
	}
}

func TestSQLiteStore_GuideRoundtripCreateUpdateGet(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "guideRoundtrip.store"
	db, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := db.CreateGuide(g)
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

	got, err := db.GetGuide(gid)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
}

func TestSQLiteStore_GetAllGuides(t *testing.T) {
	t.Parallel()
	dbSource := t.TempDir() + "GetAllGuides.store"
	sqliteStore, err := guide.OpenSQLiteStore(dbSource)
	if err != nil {
		t.Fatal(err)
	}
	guides := []guide.Guide{
		{Name: "Osaka", Coordinate: guide.Coordinate{10, 10}},
		{Name: "Istanbul", Coordinate: guide.Coordinate{10, 10}},
		{Name: "Lagos", Coordinate: guide.Coordinate{10, 10}},
	}

	for _, g := range guides {
		_, err := sqliteStore.CreateGuide(g)
		if err != nil {
			t.Fatal(err)
		}
	}

	got := sqliteStore.GetAllGuides()
	if len(got) != len(guides) {
		t.Errorf("want GetAllGuides to return %d guides, got %d", len(guides), len(got))
	}
}

func TestSqliteStore_PoiRoundtripCreateUpdateGetPoi(t *testing.T) {
	t.Parallel()

	tempDB := t.TempDir() + "poiRoundtrip.store"
	db, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := db.CreateGuide(g)
	if err != nil {
		t.Fatal(err)
	}

	poi, err := guide.NewPointOfInterest("test", gid, guide.PoiWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := db.CreatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}
	want := "testPOI"
	poi.Id = pid
	poi.Name = want
	err = db.UpdatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetPoi(pid)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
}

func TestSQLiteStore_GetAllPois(t *testing.T) {
	t.Parallel()
	dbSource := t.TempDir() + "GetAllPois.store"
	sqliteStore, err := guide.OpenSQLiteStore(dbSource)
	if err != nil {
		t.Fatal(err)
	}

	g := guide.Guide{
		Name:       "newGuide",
		Coordinate: guide.Coordinate{10, 10},
	}

	gid, err := sqliteStore.CreateGuide(g)
	if err != nil {
		t.Fatal(err)
	}
	pois := []string{"A", "B", "C"}
	for _, name := range pois {
		poi, err := guide.NewPointOfInterest(name, gid, guide.PoiWithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		sqliteStore.CreatePoi(poi)
	}
	got := sqliteStore.GetAllPois(gid)
	if len(got) != len(pois) {
		t.Errorf("want GetAllPois to return %d points of interest, got %d", len(pois), len(got))
	}
}
