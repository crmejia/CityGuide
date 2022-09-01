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

func TestMemoryStore_GuideRoundtripCreateUpdateGet(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	newGuide, err := store.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		return
	}
	want := "Tuscany"
	newGuide.Name = want
	err = store.UpdateGuide(newGuide)

	g, err := store.GetGuide(newGuide.Id)
	if err != nil {
		t.Fatal(err)
	}

	got := g.Name
	if want != got {
		t.Errorf("want guide name to be %s, got %s", want, got)
	}
}

func TestMemoryStore_CreateNewGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	want := "Tuscany"
	g, err := store.CreateGuide(want, guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	if g.Id == 0 {
		t.Errorf("0 is not a valid ID")
	}
	if _, ok := store.Guides[g.Id]; !ok {
		t.Error("want guide to be inserted into store")
	}
}

func TestMemoryStore_UpdateFailsOnNonExistingGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		return
	}
	g.Id = 400
	err = store.UpdateGuide(g)
	if err == nil {
		t.Error("want update to fail if guide does not exist")
	}
}

func TestMemoryStore_UpdateFailsOnNonSetID(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		return
	}
	g.Id = 0
	err = store.UpdateGuide(g)

	if err == nil {
		t.Error("want update to fail if guide does not exist")
	}
}

func TestMemoryStore_AllGuidesReturnsSliceOfGuides(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	want := 3
	for i := 0; i < want; i++ {
		_, err := store.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "1"))
		if err != nil {
			t.Fatal(err)
		}
	}

	allGuides := store.GetAllGuides()
	if len(allGuides) != len(store.Guides) {
		t.Error("want GetAllHabits to return a slice of habits")
	}
}

func TestMemoryStore_PoiRoundtripCreateUpdateGetPoi(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	poi, err := store.CreatePoi("test", g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	want := "testPOI"
	poi.Name = want
	err = store.UpdatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetPoi(poi.Id)
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

	g, err := store.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	pois := []string{"A", "B", "C"}
	for _, name := range pois {
		_, err := store.CreatePoi(name, g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
	}
	got := store.GetAllPois(g.Id)
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
	sqLiteStore, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g, err := sqLiteStore.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	want := "testGuide"
	g.Name = want
	g.Id = g.Id
	err = sqLiteStore.UpdateGuide(g)
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqLiteStore.GetGuide(g.Id)
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

	want := 3
	for i := 0; i < want; i++ {
		_, err := sqliteStore.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
	}

	got := sqliteStore.GetAllGuides()
	if len(got) != want {
		t.Errorf("want GetAllGuides to return %d guides, got %d", want, len(got))
	}
}

func TestSqliteStore_PoiRoundtripCreateUpdateGetPoi(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "poiRoundtrip.store"
	sqliteStore, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g, err := sqliteStore.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	poi, err := sqliteStore.CreatePoi("test", g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	want := "testPOI"
	poi.Name = want
	poi.Description = want
	err = sqliteStore.UpdatePoi(poi)
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqliteStore.GetPoi(poi.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
	if want != got.Description {
		t.Errorf("want rountrip(create,update,get) description to be %s, got %s", want, got.Description)
	}
}

func TestSQLiteStore_GetAllPois(t *testing.T) {
	t.Parallel()
	dbSource := t.TempDir() + "GetAllPois.store"
	sqliteStore, err := guide.OpenSQLiteStore(dbSource)
	if err != nil {
		t.Fatal(err)
	}

	g, err := sqliteStore.CreateGuide("newGuide", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	pois := []string{"A", "B", "C"}
	for _, name := range pois {
		_, err := sqliteStore.CreatePoi(name, g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
	}
	got := sqliteStore.GetAllPois(g.Id)
	if len(got) != len(pois) {
		t.Errorf("want GetAllPois to return %d points of interest, got %d", len(pois), len(got))
	}
}

func TestSqliteStore_GuideErrors(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "errors.db"
	sqliteStore, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sqliteStore.CreateGuide("", guide.GuideWithValidStringCoordinates("10", "10"))
	if err == nil {
		t.Error("want error on empty guide name")
	}
	_, err = sqliteStore.CreateGuide("test", guide.GuideWithValidStringCoordinates("1000", "10"))
	if err == nil {
		t.Error("want error on invalid coordinates")
	}
}

func TestSqliteStore_PoiErrors(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "errors.db"
	sqliteStore, err := guide.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sqliteStore.CreatePoi("test", 100, guide.PoiWithValidStringCoordinates("10", "10"))
	if err == nil {
		t.Error("want error on non-existing guide")
	}

	g, err := sqliteStore.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = sqliteStore.CreatePoi("", g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
	if err == nil {
		t.Error("want error on empty poi name")
	}
	_, err = sqliteStore.CreatePoi("test", g.Id, guide.PoiWithValidStringCoordinates("1110", "10"))
	if err == nil {
		t.Error("want error on invalid coordinates")
	}
}
