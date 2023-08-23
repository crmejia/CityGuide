package guide_test

import (
	"guide"
	"testing"
)

func TestOpenSQLiteStoreErrorsOnEmptyDBSource(t *testing.T) {
	t.Parallel()
	_, err := guide.OpenSQLiteStorage("")
	if err == nil {
		t.Error("Want error on empty string store source")
	}
}

func TestSQLiteStore_GetReturnsErrorOnNoGuide(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"

	sqLiteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	g, err := sqLiteStore.GetGuidebyID(99)
	if err != nil {
		t.Fatal(err)
	}
	if g != nil {
		t.Fatal("guide should be nil")
	}
}

func TestSQLiteStore_GuideRoundtripCreateUpdateGetDelete(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqLiteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g, err := guide.NewGuide(t.Name(), guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = sqLiteStore.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	want := "testGuide"
	g.Name = want
	err = sqLiteStore.UpdateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqLiteStore.GetGuidebyID(g.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}

	err = sqLiteStore.DeleteGuide(g.Id)
	if err != nil {
		t.Error("expected not error on delete")
	}

	got, err = sqLiteStore.GetGuidebyID(g.Id)
	if err != nil {
		t.Fatal(err)
	}

	if got != nil {
		t.Error("expect nil guide")
	}

}

func TestSqliteStore_DeleteGuideNoErrorsOnNopDelete(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqLiteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	err = sqLiteStore.DeleteGuide(99)
	if err != nil {
		t.Error("expected not error on delete")
	}
}

func TestSQLiteStore_GetAllGuides(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqliteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	want := 3
	for i := 0; i < want; i++ {
		g, err := guide.NewGuide("newGuide", guide.WithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		err = sqliteStore.CreateGuide(&g)
		if err != nil {
			t.Fatal(err)
		}
	}

	got := sqliteStore.GetAllGuides()
	if len(got) != want {
		t.Errorf("want GetAllGuides to return %d guides, got %d", want, len(got))
	}
}

func TestSQLiteStore_CountGuides(t *testing.T) {
	t.Parallel()
	tempDB := openTmpStorage(t)

	want := 3
	for i := 0; i < want; i++ {
		g, err := guide.NewGuide("newGuide", guide.WithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		err = tempDB.CreateGuide(&g)
		if err != nil {
			t.Fatal(err)
		}
	}

	got := tempDB.CountGuides()
	if want != got {
		t.Errorf("want CountGuides to return %d guides, got %d", want, got)
	}
}

func TestSqliteStore_PoiRoundtripCreateUpdateGetDeletePoi(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqliteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g, err := guide.NewGuide("newGuide", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = sqliteStore.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	poi, err := guide.NewPointOfInterest("test", g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
	err = sqliteStore.CreatePoi(&poi)
	if err != nil {
		t.Fatal(err)
	}
	want := "testPOI"
	poi.Name = want
	poi.Description = want
	err = sqliteStore.UpdatePoi(&poi)
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqliteStore.GetPoi(g.Id, poi.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Name {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Name)
	}
	if want != got.Description {
		t.Errorf("want rountrip(create,update,get) description to be %s, got %s", want, got.Description)
	}

	err = sqliteStore.DeletePoi(g.Id, poi.Id)
	if err != nil {
		t.Fatal(err)
	}

	got, err = sqliteStore.GetPoi(g.Id, poi.Id)
	if err != nil {
		t.Fatal(err)
	}

	if got != nil {
		t.Error("expect Poi to be nil as a result of delete")
	}
}

func TestSQLiteStore_GetAllPois(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqliteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	g, err := guide.NewGuide("newGuide", guide.WithValidStringCoordinates("10", "10"))
	err = sqliteStore.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}
	pois := []string{"A", "B", "C"}
	for _, name := range pois {
		poi, err := guide.NewPointOfInterest(name, g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
		err = sqliteStore.CreatePoi(&poi)
		if err != nil {
			t.Fatal(err)
		}
	}
	got := sqliteStore.GetAllPois(g.Id)
	if len(got) != len(pois) {
		t.Errorf("want GetAllPois to return %d points of interest, got %d", len(pois), len(got))
	}
}

func TestSqliteStore_PoiErrors(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	sqliteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	poi, err := guide.NewPointOfInterest("test", 100, guide.PoiWithValidStringCoordinates("10", "10"))
	err = sqliteStore.CreatePoi(&poi)
	if err == nil {
		t.Error("want error on non-existing guide")
	}
}

func TestSqliteStore_Search(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	input := []string{"test 1", "guide 1", "test 2"}
	for _, name := range input {
		g, err := guide.NewGuide(name, guide.WithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		err = s.CreateGuide(&g)
		if err != nil {
			t.Fatal(err)
		}
	}

	guides, err := s.Search("test")
	if err != nil {
		t.Fatal(err)
	}

	got := len(guides)
	if got != 2 {
		t.Errorf("want to have 2 search results, got %d", got)
	}

}

func TestSqliteStore_SearchNoResults(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	input := []string{"test 1", "guide 1", "test 2"}
	for _, name := range input {
		g, err := guide.NewGuide(name, guide.WithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		err = s.CreateGuide(&g)
		if err != nil {
			t.Fatal(err)
		}
	}

	guides, err := s.Search("apple")
	if err != nil {
		t.Fatal(err)
	}

	got := len(guides)
	if got != 0 {
		t.Errorf("want to have 2 search results, got %d", got)
	}

}
