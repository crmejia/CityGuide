package guide_test

import (
	"guide"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	store.Guides = map[int]guide.Guide{
		1:    guide.Guide{Id: 1, Name: "Nairobi", Coordinate: guide.Coordinate{10, 10}},
		5:    guide.Guide{Id: 5, Name: "Fukuoka", Coordinate: guide.Coordinate{11, 11}},
		2345: guide.Guide{Id: 2445, Name: "Guia de restaurantes Roma, CDMX", Coordinate: guide.Coordinate{12, 12}},
		919:  guide.Guide{Id: 919, Name: "Guia de Cuzco", Coordinate: guide.Coordinate{13, 13}},
	}
	server, err := guide.NewServer("locahost:8080", &store)
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
	g := guide.Guide{Id: 1, Name: "San Cristobal", Coordinate: guide.Coordinate{Latitude: 16.7371, Longitude: -92.6375}}
	//g.NewPointOfInterest("Cafeología", 16.737393, -92.635857)
	//g.NewPointOfInterest("Centralita Coworking", 16.739030, -92.635001)
	store.Guides = map[int]guide.Guide{
		1: g,
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
	server, err := guide.NewServer("localhost:8080", &store)
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

//func TestRunHTTPServer(t *testing.T) {
//	t.Parallel()
//	freePort, err := freeport.GetFreePort()
//	if err != nil {
//		t.Fatal(err)
//	}
//	//const (
//	//	localHostAddress = "127.0.0.1"
//	//)
//	//address := fmt.Sprintf("%s:%d", localHostAddress, freePort)
//	address := fmt.Sprintf(":%d", freePort)
//	go guide.ServerRun(address)
//
//	res, err := http.Get("http://localhost" + address)
//	for err != nil {
//		switch {
//		case strings.Contains(err.Error(), "connection refused"):
//			time.Sleep(5 * time.Millisecond)
//			res, err = http.Get(address)
//		default:
//			t.Fatal(err)
//		}
//	}
//	if res.StatusCode != http.StatusOK {
//		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
//	}
//	body, err := io.ReadAll(res.Body)
//	if err != nil {
//		t.Fatal(err)
//	}
//	want := "San Cristobal"
//	got := string(body)
//	if !strings.Contains(got, want) {
//		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
//	}
//}
