package guide_test

import (
	"guide"
	"testing"
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
