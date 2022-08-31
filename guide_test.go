package guide_test

import (
	"guide"
	"testing"
)

func TestGuideWithValidStringCoordinatesErrorsOnInvalidCoordinates(t *testing.T) {
	t.Parallel()
	store := guide.OpenMemoryStore()
	//Latitude valid range [-90,90]
	//Longitude valid range [-180,180]
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
		_, err := store.CreateGuide(tc.name, guide.GuideWithValidStringCoordinates(tc.latitude, tc.longitude))
		if err == nil {
			t.Errorf("want error on %s, Latitude: %s  Longitude: %s", tc.name, tc.latitude, tc.longitude)
		}
	}
}
