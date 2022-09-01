package guide_test

import (
	"fmt"
	"github.com/phayes/freeport"
	"guide"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServerErrors(t *testing.T) {
	t.Parallel()
	s := guide.OpenMemoryStore()
	_, err := guide.NewServer("", &s)
	if err == nil {
		t.Errorf("want error on empty server address")
	}

	_, err = guide.NewServer("address", nil)
	if err == nil {
		t.Errorf("want error on nil store")
	}
}

func TestIndexHandler(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	store.CreateGuide("Nairobi", guide.GuideWithValidStringCoordinates("10", "10"))
	store.CreateGuide("Fukuoka", guide.GuideWithValidStringCoordinates("10", "10"))
	store.CreateGuide("Guia de Restaurates Roma, CDMX", guide.GuideWithValidStringCoordinates("10", "10"))

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := server.HandleIndex()
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

func TestGuideHandlerRendersMap(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/1", nil)
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
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/1", nil)
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
	store := guide.OpenMemoryStore()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/", nil)
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
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/create", nil)
	handler := server.HandleCreateGuide()
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

func TestCreateGuideHandlerPostCreatesGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	form := strings.NewReader("name=Test&description=blah blah&latitude=10&longitude=10")
	req := httptest.NewRequest(http.MethodPost, "/guide/create", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := server.HandleCreateGuide()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}
	if len(store.Guides) != 1 {
		t.Error("want store to contain new guide")
	}
}

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
	store := guide.OpenMemoryStore()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreateGuide()
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

func TestCreatePoiHandlerGetRendersForm(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.GuideWithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/guide/poi/create/1", nil)
	handler := server.HandleCreatePoi()
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

func TestCreatePoiHandlerErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		url        string
		statusCode int
		want       string
	}{
		{"/guide/poi/create", http.StatusBadRequest, "no guideid provided"},
		{"/guide/poi/create/one", http.StatusBadRequest, "please provide valid guide id"},
		{"/guide/poi/create/1", http.StatusNotFound, "guide not found"},
	}
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreatePoi()
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
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("test", guide.GuideWithValidStringCoordinates("10", "10"))

	server, err := guide.NewServer("localhost:8080", &store)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.HandleCreatePoi()
	for _, tc := range testCases {
		rec := httptest.NewRecorder()

		target := fmt.Sprintf("/guide/poi/create/%d", g.Id)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(tc.form))
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

func TestCreatePoiHandlerPost(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("San Cristobal", guide.GuideWithValidStringCoordinates("16.7371", "-92.6375"))
	if err != nil {
		t.Fatal(err)
	}

	freeport, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freeport)
	server, err := guide.NewServer(address, &store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	form := strings.NewReader("name=Test&description=blah blah&latitude=10&longitude=10")
	target := fmt.Sprintf("/guide/poi/create/%d", g.Id)
	req := httptest.NewRequest(http.MethodPost, target, form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := server.HandleCreatePoi()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}
	pois := store.GetAllPois(g.Id)
	if len(pois) != 1 {
		t.Error("want store to contain new poi")
	}

	got := pois[0]
	if got.Description != "blah blah" {
		t.Error("want poi description to be set")
	}
}
