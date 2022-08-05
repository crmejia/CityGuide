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
		t.Errorf("name a guide named %s, got %s", name, got)
	}
}

func TestNewGuideNameCannotBeEmpty(t *testing.T) {
	t.Parallel()
	_, err := guide.NewGuide("")
	if err == nil {
		t.Error("want error on empty name")
	}
}

func TestWithValidCoordinates(t *testing.T) {
	t.Parallel()
	//latitude valid range [-90,90]
	//long valid range [-180,180]
	testCases := []struct {
		name                string
		latitude, longitude float64
	}{
		{name: "invalid latitude", latitude: -91, longitude: 100},
		{name: "invalid latitude", latitude: 92, longitude: 100},
		{name: "invalid longitude", latitude: -41, longitude: 181},
		{name: "invalid longitude", latitude: 1, longitude: -180.1},
	}
	for _, tc := range testCases {
		_, err := guide.NewGuide("test", guide.WithValidCoordinates(tc.latitude, tc.longitude))
		if err == nil {
			t.Errorf("want error on %s, latitude: %f  longitude: %f", tc.name, tc.latitude, tc.longitude)
		}
	}
}

func TestIndexHandlerRendersMap(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	guide.IndexHandler(recorder, req)

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
