package guide

import (
	_ "embed"
	"errors"
	"net/http"
	"strconv"
)

type Guide struct {
	Id          int
	Name        string
	Description string
	Coordinate  Coordinate
	Pois        []pointOfInterest
}

type pointOfInterest struct {
	Id         int
	Coordinate Coordinate
	Name       string
}

type Coordinate struct {
	Latitude, Longitude float64
}

// TODO Make private
func NewCoordinate(latitude, longitude float64) (Coordinate, error) {
	if latitude < -90 || latitude > 90 {
		return Coordinate{}, errors.New("latitude has to be in the -90°, 90° range")
	}
	if longitude < -180 || longitude > 180 {
		return Coordinate{}, errors.New("longitude has to be in the -90°, 90° range")
	}
	return Coordinate{Latitude: latitude, Longitude: longitude}, nil
}

type guideOption func(*Guide) error

func WithValidCoordinates(latitude, longitude float64) guideOption {
	return func(g *Guide) error {
		coord, err := NewCoordinate(latitude, longitude)
		if err != nil {
			return err
		}
		g.Coordinate = coord
		return nil

	}
}

func WithValidStringCoordinates(latitude, longitude string) guideOption {
	return func(g *Guide) error {
		if latitude == "" {
			return errors.New("latitude cannot be empty")
		}
		if longitude == "" {
			return errors.New("longitude cannot be empty")
		}

		lat, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			return errors.New("latitude has to be a number")
		}
		lon, err := strconv.ParseFloat(longitude, 64)
		if err != nil {
			return errors.New("longitude hast to be a number")
		}
		coord, err := NewCoordinate(lat, lon)
		if err != nil {
			return err
		}
		g.Coordinate = coord
		return nil

	}
}
func WithDescription(description string) guideOption {
	return func(g *Guide) error {
		g.Description = description
		return nil
	}
}

//WithLocation
//WithLocationLookUp

//AreSpots outOfGuideBounds

func NewGuide(name string, opts ...guideOption) (Guide, error) {
	if name == "" {
		return Guide{}, errors.New("guide name cannot be empty")
	}
	g := Guide{
		Name: name,
		Pois: []pointOfInterest{},
	}

	for _, opt := range opts {
		err := opt(&g)
		if err != nil {
			return Guide{}, err
		}
	}
	return g, nil
}

func (g *Guide) NewPointOfInterest(name string, latitude float64, longitude float64) error {
	coordinates, err := NewCoordinate(latitude, longitude)
	if err != nil {
		return err
	}
	poi := pointOfInterest{
		Coordinate: coordinates,
		Name:       name,
	}
	g.Pois = append(g.Pois, poi)
	return nil
}

func validateGuideForm(w http.ResponseWriter, r *http.Request) *Guide {
	form := struct {
		Name, Description, Latitude, Longitude string
		Errors                                 []string
	}{
		Name:        r.PostFormValue("name"),
		Description: r.PostFormValue("description"),
		Latitude:    r.PostFormValue("latitude"),
		Longitude:   r.PostFormValue("longitude"),
	}

	g, err := NewGuide(form.Name, WithValidStringCoordinates(form.Latitude, form.Longitude), WithDescription(form.Description))
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		render(w, r, "templates/createGuide.html", form)
		return nil
	}
	return &g
}
