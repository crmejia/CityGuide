package guide_test

import (
	"guide"
	"testing"
)

func TestNewGuideErrors(t *testing.T) {
	testCases := []struct {
		name                string
		latitude, longitude string
	}{
		{name: "invalid Latitude", latitude: "-91", longitude: "100"},
		{name: "invalid Latitude", latitude: "92", longitude: "100"},
		{name: "invalid Longitude", latitude: "-41", longitude: "181"},
		{name: "invalid Longitude", latitude: "1", longitude: "-180.1"},
		{name: "empty Latitude", latitude: "", longitude: "100"},
		{name: "empty Longitude", latitude: "11", longitude: ""},
	}
	for _, tc := range testCases {
		_, err := guide.NewGuide(tc.name, guide.WithValidStringCoordinates(tc.latitude, tc.longitude))
		if err == nil {
			t.Errorf("want error on %s, Latitude: %s  Longitude: %s", tc.name, tc.latitude, tc.longitude)
		}
	}
	t.Parallel()
}
func TestNewGuideErrorOnEmptyName(t *testing.T) {
	_, err := guide.NewGuide("", guide.WithValidStringCoordinates("10", "10"))
	if err == nil {
		t.Error("want error if name is empty")
	}
}

func TestNewPoiErrors(t *testing.T) {
	testCases := []struct {
		name                string
		latitude, longitude string
	}{
		{name: "invalid Latitude", latitude: "-91", longitude: "100"},
		{name: "invalid Latitude", latitude: "92", longitude: "100"},
		{name: "invalid Longitude", latitude: "-41", longitude: "181"},
		{name: "invalid Longitude", latitude: "1", longitude: "-180.1"},
		{name: "empty Latitude", latitude: "", longitude: "100"},
		{name: "empty Longitude", latitude: "11", longitude: ""},
	}
	//g, err := guide.NewGuide("test", guide.WithValidStringCoordinates("10", "10"))
	//if err != nil {
	//	t.Fatal(err)
	//}
	for _, tc := range testCases {
		_, err := guide.NewPointOfInterest(tc.name, 1, guide.PoiWithValidStringCoordinates(tc.latitude, tc.longitude))
		if err == nil {
			t.Errorf("want error on %s, Latitude: %s  Longitude: %s", tc.name, tc.latitude, tc.longitude)
		}
	}
	t.Parallel()
}

func TestNewPoiErrorOnEmptyName(t *testing.T) {
	_, err := guide.NewPointOfInterest("", 1, guide.PoiWithValidStringCoordinates("10", "10"))
	if err == nil {
		t.Error("want error if name is empty")
	}
}
