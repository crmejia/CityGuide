package guide_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/phayes/freeport"
	"guide"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestNewServerErrors(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	_, err := guide.NewServer("", s, os.Stdout)
	if err == nil {
		t.Errorf("want error on empty server address")
	}

	_, err = guide.NewServer("address", nil, os.Stdout)
	if err == nil {
		t.Errorf("want error on nil store")
	}
}

func TestIndexHandler(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("Nairobi", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	//err = s.CreateGuide("Fukuoka", guide.WithValidStringCoordinates("10", "10"))
	//err = s.CreateGuide("Guia de Restaurates Roma, CDMX", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := server.HandleGuides()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("want status 200 OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "Nairobi"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s, got:\n%s", want, got)
	}
}

func TestGetIndexRoute(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		path               string
		expectedStatusCode int
	}{
		{"/", http.StatusOK}, //should be StatusFound 302 but httptest.Client follows redirects
		{"/guides", http.StatusOK},
		{"/unknownroute", http.StatusNotFound},
	}

	s := openTmpStorage(t)
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(server.Routes())
	defer ts.Close()

	client := ts.Client()
	for _, tc := range testCases {
		res, err := client.Get(ts.URL + tc.path)

		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != tc.expectedStatusCode {
			t.Errorf("for path %s want status %d OK, got %d", tc.path, tc.expectedStatusCode, res.StatusCode)
		}
	}
}

func TestGetGuideRoute(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		path               string
		expectedStatusCode int
	}{
		{"/guide/1", http.StatusOK}, //should be StatusFound 302 but httptest.Client follows redirects
		{"/guide/2", http.StatusNotFound},
	}

	s := openTmpStorage(t)

	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(server.Routes())
	defer ts.Close()

	client := ts.Client()
	for _, tc := range testCases {
		res, err := client.Get(ts.URL + tc.path)

		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != tc.expectedStatusCode {
			t.Errorf("for path %s want status %d OK, got %d", tc.path, tc.expectedStatusCode, res.StatusCode)
		}
	}
}
func TestGuideHandlerRendersMap(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
	handler := server.HandleGuide()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "San Cristobal"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

func TestGuideHandlerRenders404NotFound(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
	handler := server.HandleGuide()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "Not Found"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

func TestGuideHandlerRenders400NoId(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "parse"})
	handler := server.HandleGuide()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "not able to parse"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

func TestCreateGuideHandlerGetRendersForm(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/create", nil)
	handler := server.HandleCreateGuideGet()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "Create New Guide"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

//todo replace s.Guides with count
//func TestCreateGuideHandlerPostCreatesGuide(t *testing.T) {
//	t.Parallel()
//	s := openTmpStorage(t)
//	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
//	if err != nil {
//		t.Fatal(err)
//	}
//	rec := httptest.NewRecorder()
//	form := strings.NewReader("name=Test&description=blah blah&latitude=10&longitude=10")
//	req := httptest.NewRequest(http.MethodPost, "/guide/create", form)
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//	handler := server.HandleCreateGuidePost()
//	handler(rec, req)
//
//	res := rec.Result()
//	if res.StatusCode != http.StatusSeeOther {
//		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
//	}
//	if len(s.Guides) != 1 {
//		t.Error("want store to contain new guide")
//	}
//
//	g, err := s.GetGuidebyID(s.NextGuideKey - 1)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if g.Description == "" {
//		t.Error("want guide description to not be empty")
//	}
//}

func TestCreateGuideHandlerPostFormErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		form string
		want string
	}{
		{"name=&description=blah blah&latitude=10&longitude=10", "name cannot be empty"},
		{"name=test&latitude=&longitude=10", "latitude cannot be empty"},
		{"name=test&latitude=10&longitude=", "longitude cannot be empty"},
		{"name=test&latitude=notanumber&longitude=10", "latitude has to be a number"},
		{"name=test&latitude=10&longitude=notanumber", "longitude has to be a number"},
	}
	s := openTmpStorage(t)
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreateGuidePost()
	for _, tc := range testCases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/guide/create", strings.NewReader(tc.form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler(rec, req)

		res := rec.Result()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		got := string(body)
		if !strings.Contains(got, tc.want) {
			t.Errorf("want index to contain %s\nGot:\n%s", tc.want, got)
		}
	}
}

func TestEditGuideRoute(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		path               string
		expectedStatusCode int
	}{
		{"/guide/1/edit", http.StatusOK}, //should be StatusFound 302 but httptest.Client follows redirects
		{"/guide/2/edit", http.StatusNotFound},
	}

	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(server.Routes())
	defer ts.Close()

	client := ts.Client()
	for _, tc := range testCases {
		res, err := client.Get(ts.URL + tc.path)

		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != tc.expectedStatusCode {
			t.Errorf("for path %s want status %d OK, got %d", tc.path, tc.expectedStatusCode, res.StatusCode)
		}
	}
}

func TestDeleteGuideHandlerDeletesGuide(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
	handler := server.HandleDeleteGuide()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, res.StatusCode)
	}

	if len(s.GetAllGuides()) != 0 {
		t.Error("expected table to be empty after delete")
	}
}

func TestPoiHandlerRendersView(t *testing.T) {
	t.Parallel()
	server := newProvisionedServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"guideID": "1",
		"poiID":   "1",
	})

	handler := server.HandlePoi()
	handler(rec, req)

	result := rec.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, body %d", result.StatusCode)
	}
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "test 1"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want result to contain %s\nGot:\n%s", want, got)

	}
}

func TestCreatePoiHandlerGetRendersForm(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}
	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/poi/create/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
	handler := server.HandleCreatePoiGet()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "Add Point of Interest to"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

// TODO {"/guide/poi/create/1", http.StatusNotFound, "guide not found"}, test poi create on non-existing guide id
func TestCreatePoiHandlerErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		url        string
		statusCode int
		want       string
	}{
		{"/guide/poi/create", http.StatusBadRequest, "no guideid provided"},
		{"/guide/poi/create/one", http.StatusBadRequest, "no guideid provided"},
		//{"/guide/poi/create/1", http.StatusNotFound, "guide not found"},
	}
	s := openTmpStorage(t)
	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreatePoiPost()
	for _, tc := range testCases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.url, nil)
		handler(rec, req)

		res := rec.Result()
		if res.StatusCode != tc.statusCode {
			t.Errorf("expected status %d, got %d", tc.statusCode, res.StatusCode)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		got := string(body)
		if !strings.Contains(got, tc.want) {
			t.Errorf("want index to contain %s\nGot:\n%s", tc.want, got)
		}
	}
}

func TestCreatePoiHandlerFormErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		form string
		want string
	}{
		{"name=&description=blah blah&latitude=10&longitude=10", "name cannot be empty"},
		{"name=test&latitude=&longitude=10", "latitude cannot be empty"},
		{"name=test&latitude=10&longitude=", "longitude cannot be empty"},
		{"name=test&latitude=notanumber&longitude=10", "latitude has to be a number"},
		{"name=test&latitude=10&longitude=notanumber", "longitude has to be a number"},
	}
	s := openTmpStorage(t)
	g, err := guide.NewGuide("test", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreatePoiPost()
	for _, tc := range testCases {
		rec := httptest.NewRecorder()

		target := fmt.Sprintf("/guide/poi/create/%d", g.Id)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(tc.form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
		handler(rec, req)

		res := rec.Result()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		got := string(body)
		if !strings.Contains(got, tc.want) {
			t.Errorf("want index to contain %s\nGot:\n%s", tc.want, got)
		}
	}
}

func TestCreatePoiHandlerPost(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("16.7371", "-92.6375"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	form := strings.NewReader("name=Test&description=blah blah&latitude=10&longitude=10")
	target := fmt.Sprintf("/guide/poi/create/%d", g.Id)
	req := httptest.NewRequest(http.MethodPost, target, form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(1, 10)})
	handler := server.HandleCreatePoiPost()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}
	pois := s.GetAllPois(g.Id)
	if len(pois) != 1 {
		t.Error("want store to contain new poi")
	}

	got := pois[0]
	if got.Description != "blah blah" {
		t.Error("want poi description to be set")
	}
}

func TestDeletePoiHandlerDeletesPoi(t *testing.T) {
	t.Parallel()
	s := openTmpStorage(t)
	g, err := guide.NewGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreateGuide(&g)
	if err != nil {
		t.Fatal(err)
	}

	poi, err := guide.NewPointOfInterest("test", g.Id, guide.PoiWithValidStringCoordinates("10", "1"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.CreatePoi(&poi)
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", s, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = mux.SetURLVars(req,
		map[string]string{
			"id":      strconv.FormatInt(poi.Id, 10),
			"guideID": strconv.FormatInt(g.Id, 10),
		})
	handler := server.HandleDeletePoi()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, res.StatusCode)
	}

	if len(s.GetAllPois(g.Id)) != 0 {
		t.Error("expected poi table to be empty after delete")
	}
}

func TestSearchGuide(t *testing.T) {
	t.Parallel()
	server := newProvisionedServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guide/search?q=guide", nil)
	//req = mux.SetURLVars(req, map[string]string{"q": "guide"})
	handler := server.HandleGuides()
	handler(rec, req)

	res := rec.Result()
	if http.StatusOK != res.StatusCode {
		t.Errorf("want status code %d, got %d", http.StatusOK, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	want := "guide 1"
	if !strings.Contains(got, want) {
		t.Errorf("want body to contain %s, got %s instead.", want, got)
	}
}

func TestSearchGuideEmptySearchReturnsAll(t *testing.T) {
	t.Parallel()
	server := newProvisionedServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guide/search", nil)

	handler := server.HandleGuides()
	handler(rec, req)

	res := rec.Result()
	if http.StatusOK != res.StatusCode {
		t.Errorf("want status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	want := "test 1"
	if !strings.Contains(got, want) {
		t.Errorf("want body to contain %s, got %s instead.", want, got)
	}
}

func TestSearchGuideNoMatchReturnsNothing(t *testing.T) {
	t.Parallel()
	server := newProvisionedServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guide/search?q=apple", nil)

	handler := server.HandleGuides()
	handler(rec, req)

	res := rec.Result()
	if http.StatusNotFound != res.StatusCode {
		t.Errorf("want status code %d, got %d", http.StatusNotFound, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	want := "no guide found"
	if !strings.Contains(got, want) {
		t.Errorf("want body to contain %s, got %s instead.", want, got)
	}
}

func TestServer_HandleGuideCount(t *testing.T) {
	t.Parallel()
	server := newProvisionedServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guide/count", nil)

	handler := server.HandleGuideCount()
	handler(rec, req)

	res := rec.Result()
	if http.StatusOK != res.StatusCode {
		t.Errorf("want status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	want := "3"
	if !strings.Contains(got, want) {
		t.Errorf("want body to contain %s, got %s instead.", want, got)
	}
}

// test helpers
func openTmpStorage(t *testing.T) guide.Storage {
	tempDB := t.TempDir() + t.Name() + ".store"
	// I don't have to close the db connection because the clients will after running query?
	//i'm not sure it needs to be closed here. If deferred it will get destroyed.
	sqliteStore, err := guide.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	return sqliteStore
}

func newProvisionedServer(t *testing.T) *guide.Server {
	storage := openTmpStorage(t)
	input := []string{"test 1", "guide 1", "test 2"}
	for _, guideName := range input {
		g, err := guide.NewGuide(guideName, guide.WithValidStringCoordinates("10", "10"))
		if err != nil {
			t.Fatal(err)
		}
		err = storage.CreateGuide(&g)
		if err != nil {
			t.Fatal(err)
		}
		for _, poiName := range input {
			p, err := guide.NewPointOfInterest(poiName, g.Id, guide.PoiWithValidStringCoordinates("10", "10"))
			if err != nil {
				t.Fatal(err)
			}

			err = storage.CreatePoi(&p)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, storage, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	return &server

}
