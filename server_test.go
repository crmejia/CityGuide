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
	s := guide.OpenMemoryStore()
	_, err := guide.NewServer("", &s, os.Stdout)
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
	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("Nairobi", guide.WithValidStringCoordinates("10", "10"))
	_, err = store.CreateGuide("Fukuoka", guide.WithValidStringCoordinates("10", "10"))
	_, err = store.CreateGuide("Guia de Restaurates Roma, CDMX", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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

	store := guide.OpenMemoryStore()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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

	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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

func TestCreateGuideHandlerPostCreatesGuide(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	form := strings.NewReader("name=Test&description=blah blah&latitude=10&longitude=10")
	req := httptest.NewRequest(http.MethodPost, "/guide/create", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := server.HandleCreateGuidePost()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}
	if len(store.Guides) != 1 {
		t.Error("want store to contain new guide")
	}

	g, err := store.GetGuide(store.NextGuideKey - 1)
	if err != nil {
		t.Fatal(err)
	}
	if g.Description == "" {
		t.Error("want guide description to not be empty")
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
	server, err := guide.NewServer(address, &store, os.Stdout)
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

	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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
func TestCreatePoiHandlerGetRendersForm(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	_, err := store.CreateGuide("San Cristobal", guide.WithValidStringCoordinates("10", "10"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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

// TODO {"/guide/poi/create/1", http.StatusNotFound, "guide not found"},
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
	store := guide.OpenMemoryStore()
	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("test", guide.WithValidStringCoordinates("10", "10"))

	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
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
	store := guide.OpenMemoryStore()
	g, err := store.CreateGuide("San Cristobal", guide.WithValidStringCoordinates("16.7371", "-92.6375"))
	if err != nil {
		t.Fatal(err)
	}

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	server, err := guide.NewServer(address, &store, os.Stdout)
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
	pois := store.GetAllPois(g.Id)
	if len(pois) != 1 {
		t.Error("want store to contain new poi")
	}

	got := pois[0]
	if got.Description != "blah blah" {
		t.Error("want poi description to be set")
	}
}

//func TestCreateUserHandlerGetRendersForm(t *testing.T) {
//	t.Parallel()
//	store := guide.OpenMemoryStore()
//	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
//	if err != nil {
//		t.Fatal(err)
//	}
//	rec := httptest.NewRecorder()
//	req := httptest.NewRequest(http.MethodGet, "/user/signup", nil)
//	handler := server.HandleCreateUser()
//	handler(rec, req)
//
//	res := rec.Result()
//	if res.StatusCode != http.StatusOK {
//		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
//	}
//	body, err := io.ReadAll(res.Body)
//	if err != nil {
//		t.Fatal(err)
//	}
//	want := "Create your account"
//	got := string(body)
//	if !strings.Contains(got, want) {
//		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
//	}
//}
//
//func TestCreateUserHandlerPostCreatesUser(t *testing.T) {
//	t.Parallel()
//	store := guide.OpenMemoryStore()
//	server, err := guide.NewServer("localhost:8080", &store, os.Stdout)
//	if err != nil {
//		t.Fatal(err)
//	}
//	rec := httptest.NewRecorder()
//	form := strings.NewReader("username=test&password=password&confirm-password=password&email=email@test.com")
//	req := httptest.NewRequest(http.MethodPost, "/user/create", form)
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//	handler := server.HandleCreateUser()
//	handler(rec, req)
//
//	res := rec.Result()
//	if res.StatusCode != http.StatusSeeOther {
//		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
//	}
//	if len(store.Users) != 1 {
//		t.Error("want store to contain new user")
//	}
//
//	u, err := store.GetUser(store.NextUserKey - 1)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if u.Username == "" {
//		t.Error("want guide description to not be empty")
//	}
//}
//
//func TestCreateUserHandlerPostFormErrors(t *testing.T) {
//	t.Parallel()
//	testCases := []struct {
//		form string
//		want string
//	}{
//		{"username=&password=password&confirm-password=password&email=email@test.com", "username cannot be empty"},
//		{"username=test&password=&confirm-password=password&email=email@test.com", "password cannot be empty"},
//		{"username=test&password=short&confirm-password=short&email=email@test.com", "password has to be at least 8 characters long"},
//		{"username=test&password=password&confirm-password=passsentence&email=email@test.com", "passwords do not match"},
//		{"username=test&password=password&confirm-password=password&email=", "email cannot be empty"},
//		{"username=test&password=password&confirm-password=password&email=email", "email has to be a valid address"},
//	}
//	store := guide.OpenMemoryStore()
//	freePort, err := freeport.GetFreePort()
//	if err != nil {
//		t.Fatal(err)
//	}
//	address := fmt.Sprintf("localhost:%d", freePort)
//	server, err := guide.NewServer(address, &store, os.Stdout)
//	if err != nil {
//		t.Fatal(err)
//	}
//	handler := server.HandleCreateUser()
//	for _, tc := range testCases {
//		rec := httptest.NewRecorder()
//		req := httptest.NewRequest(http.MethodPost, "/user/create", strings.NewReader(tc.form))
//		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//		handler(rec, req)
//
//		res := rec.Result()
//		if res.StatusCode != http.StatusBadRequest {
//			t.Errorf("expected status 400 bad request, got %d", res.StatusCode)
//		}
//		body, err := io.ReadAll(res.Body)
//		if err != nil {
//			t.Fatal(err)
//		}
//		got := string(body)
//		if !strings.Contains(got, tc.want) {
//			t.Errorf("want page to contain %s\nGot:\n%s", tc.want, got)
//		}
//	}
//}
