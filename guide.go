package guide

import (
	_ "embed"
	"errors"
	"net/http"
	"strconv"
)

type Guide struct {
	Id          int64
	Name        string
	Description string
	Coordinate  Coordinate
	Pois        *[]pointOfInterest
}

type Coordinate struct {
	Latitude, Longitude float64
}

func newCoordinate(latitude, longitude float64) (Coordinate, error) {
	if latitude < -90 || latitude > 90 {
		return Coordinate{}, errors.New("latitude has to be in the -90°, 90° range")
	}
	if longitude < -180 || longitude > 180 {
		return Coordinate{}, errors.New("longitude has to be in the -90°, 90° range")
	}
	return Coordinate{Latitude: latitude, Longitude: longitude}, nil
}

type guideOption func(*Guide) error

func GuideWithValidStringCoordinates(latitude, longitude string) guideOption {
	return func(g *Guide) error {
		coordinate, err := parseCoordinates(latitude, longitude)
		if err != nil {
			return err
		}
		g.Coordinate = coordinate
		return nil

	}
}

func GuideWithDescription(description string) guideOption {
	return func(g *Guide) error {
		g.Description = description
		return nil
	}
}

func NewGuide(name string, opts ...guideOption) (Guide, error) {
	if name == "" {
		return Guide{}, errors.New("guide name cannot be empty")
	}
	g := Guide{
		Name: name,
		Pois: &[]pointOfInterest{},
	}

	for _, opt := range opts {
		err := opt(&g)
		if err != nil {
			return Guide{}, err
		}
	}
	return g, nil
}

type pointOfInterest struct {
	Id          int64
	GuideID     int64
	Coordinate  Coordinate
	Name        string
	Description string
}

type poiOption func(*pointOfInterest) error

func PoiWithValidStringCoordinates(latitude, longitude string) poiOption {
	return func(poi *pointOfInterest) error {
		coordinate, error := parseCoordinates(latitude, longitude)
		if error != nil {
			return error
		}
		poi.Coordinate = coordinate
		return nil
	}
}

func PoiWithDescription(description string) poiOption {
	return func(poi *pointOfInterest) error {
		poi.Description = description
		return nil
	}
}

func NewPointOfInterest(name string, guideID int64, opts ...poiOption) (pointOfInterest, error) {
	if name == "" {
		return pointOfInterest{}, errors.New("poi name cannot be empty")
	}
	if guideID <= 0 {
		return pointOfInterest{}, errors.New("guide ID cannot be empty")
	}
	poi := pointOfInterest{
		Name:    name,
		GuideID: guideID,
	}

	for _, opt := range opts {
		err := opt(&poi)
		if err != nil {
			return pointOfInterest{}, err
		}
	}
	return poi, nil
}

func newGuideForm(w http.ResponseWriter, r *http.Request) *Guide {
	form := struct {
		Name, Description, Latitude, Longitude string
		Errors                                 []string
	}{
		Name:        r.PostFormValue("name"),
		Description: r.PostFormValue("description"),
		Latitude:    r.PostFormValue("latitude"),
		Longitude:   r.PostFormValue("longitude"),
	}

	g, err := NewGuide(form.Name, GuideWithValidStringCoordinates(form.Latitude, form.Longitude), GuideWithDescription(form.Description))
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, "templates/createGuide.html", form)
		return nil
	}
	return &g
}

type poiForm struct {
	GuideID                                int64
	GuideName                              string
	Name, Description, Latitude, Longitude string
	Errors                                 []string
}

func parseCoordinates(latitude, longitude string) (Coordinate, error) {
	if latitude == "" {
		return Coordinate{}, errors.New("latitude cannot be empty")
	}
	if longitude == "" {
		return Coordinate{}, errors.New("longitude cannot be empty")
	}

	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return Coordinate{}, errors.New("latitude has to be a number")
	}
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return Coordinate{}, errors.New("longitude has to be a number")
	}
	coord, err := newCoordinate(lat, lon)
	if err != nil {
		return Coordinate{}, err
	}
	return coord, nil
}
