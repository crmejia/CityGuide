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
	store := guide.MemoryStore{
		Guides: map[guide.Coordinate]guide.Guide{
			guide.Coordinate{10, 10}: guide.Guide{Name: "Nairobi", Coordinate: guide.Coordinate{10, 10}},
			guide.Coordinate{11, 11}: guide.Guide{Name: "Fukuoka", Coordinate: guide.Coordinate{11, 11}},
			guide.Coordinate{12, 12}: guide.Guide{Name: "Guia de restaurantes Roma, CDMX", Coordinate: guide.Coordinate{12, 12}},
			guide.Coordinate{13, 13}: guide.Guide{Name: "Guia de Cuzco", Coordinate: guide.Coordinate{13, 13}},
		}}
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

//func TestGuideHandlerRendersMap(t *testing.T) {
//	t.Parallel()
//	g, err := guide.NewGuide("San Cristobal", guide.WithValidCoordinates(16.7371, -92.6375))
//	if err != nil {
//		t.Fatal(err)
//	}
//	g.NewPointOfInterest("Cafeología", 16.737393, -92.635857)
//	g.NewPointOfInterest("Centralita Coworking", 16.739030, -92.635001)
//	recorder := httptest.NewRecorder()
//	req := httptest.NewRequest(http.MethodGet, "/?guideid=1", nil)
//	controller := guide.Server{
//		Guides: map[int]guide.Guide{1: g},
//	}
//	controller.GuideHandler(recorder, req)
//	//guide.IndexHandler(recorder, req)
//
//	res := recorder.Result()
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
