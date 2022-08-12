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
	"time"
)

func TestNewGuide(t *testing.T) {
	t.Parallel()
	name := "San Cristobal de las Casas"
	//16.737126144865194, -92.63750443638673
	lat := float64(16.73)
	lon := float64(-92.63)
	g, err := guide.NewGuide(name, guide.WithValidCoordinates(lat, lon))
	if err != nil {
		t.Fatal(err)
	}

	got := g.Name

	if name != got {
		t.Errorf("Name a Guide named %s, got %s", name, got)
	}
}

func TestNewGuideNameCannotBeEmpty(t *testing.T) {
	t.Parallel()
	_, err := guide.NewGuide("")
	if err == nil {
		t.Error("want error on empty Name")
	}
}

func TestWithValidCoordinatesErrorsOnInvalidCoordinates(t *testing.T) {
	t.Parallel()
	//Latitude valid range [-90,90]
	//Longitude valid range [-180,180]
	testCases := []struct {
		name                string
		latitude, longitude float64
	}{
		{name: "invalid Latitude", latitude: -91, longitude: 100},
		{name: "invalid Latitude", latitude: 92, longitude: 100},
		{name: "invalid Longitude", latitude: -41, longitude: 181},
		{name: "invalid Longitude", latitude: 1, longitude: -180.1},
	}
	for _, tc := range testCases {
		_, err := guide.NewGuide("test", guide.WithValidCoordinates(tc.latitude, tc.longitude))
		if err == nil {
			t.Errorf("want error on %s, Latitude: %f  Longitude: %f", tc.name, tc.latitude, tc.longitude)
		}
	}
}

func TestIndexHandlerRendersMap(t *testing.T) {
	t.Parallel()
	g, err := guide.NewGuide("San Cristobal", guide.WithValidCoordinates(16.7371, -92.6375))
	if err != nil {
		t.Fatal(err)
	}
	g.NewPointOfInterest("Cafeología", 16.737393, -92.635857)
	g.NewPointOfInterest("Centralita Coworking", 16.739030, -92.635001)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?guideid=1", nil)
	controller := guide.Controller{
		Guides: map[int]guide.Guide{1: g},
	}
	controller.GuideHandler(recorder, req)
	//guide.IndexHandler(recorder, req)

	res := recorder.Result()
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

func TestRunHTTPServer(t *testing.T) {
	t.Parallel()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	//const (
	//	localHostAddress = "127.0.0.1"
	//)
	//address := fmt.Sprintf("%s:%d", localHostAddress, freePort)
	address := fmt.Sprintf(":%d", freePort)
	go guide.ServerRun(address)

	res, err := http.Get("http://localhost" + address)
	for err != nil {
		switch {
		case strings.Contains(err.Error(), "connection refused"):
			time.Sleep(5 * time.Millisecond)
			res, err = http.Get(address)
		default:
			t.Fatal(err)
		}
	}
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
