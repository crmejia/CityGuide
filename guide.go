package guide

import (
	_ "embed"
	"errors"
	"strconv"
)

func WithValidStringCoordinates(latitude, longitude string) guideOption {
	return func(g *guide) error {
		coordinate, err := parseCoordinates(latitude, longitude)
		if err != nil {
			return err
		}
		g.Coordinate = coordinate
		return nil

	}
}

func WithDescription(description string) guideOption {
	return func(g *guide) error {
		g.Description = description
		return nil
	}
}

func PoiWithValidStringCoordinates(latitude, longitude string) poiOption {
	return func(poi *pointOfInterest) error {
		coordinate, err := parseCoordinates(latitude, longitude)
		if err != nil {
			return err
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

type guide struct {
	Id          int64
	Name        string
	Description string
	Coordinate  coordinate
	Pois        []pointOfInterest

	// guide.mapArea/coordinates}
}

type coordinate struct {
	Latitude, Longitude float64
}

//todo type boundedCoordinate coordinate

func newCoordinate(latitude, longitude float64) (coordinate, error) {
	if latitude < -90 || latitude > 90 {
		return coordinate{}, errors.New("latitude has to be in the -90°, 90° range")
	}
	if longitude < -180 || longitude > 180 {
		return coordinate{}, errors.New("longitude has to be in the -90°, 90° range")
	}
	return coordinate{Latitude: latitude, Longitude: longitude}, nil
}

func parseCoordinates(latitude, longitude string) (coordinate, error) {
	if latitude == "" {
		return coordinate{}, errors.New("latitude cannot be empty")
	}
	if longitude == "" {
		return coordinate{}, errors.New("longitude cannot be empty")
	}

	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return coordinate{}, errors.New("latitude has to be a number")
	}
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return coordinate{}, errors.New("longitude has to be a number")
	}
	coord, err := newCoordinate(lat, lon)
	if err != nil {
		return coordinate{}, err
	}
	return coord, nil
}

type guideOption func(*guide) error

func NewGuide(name string, opts ...guideOption) (guide, error) {
	if name == "" {
		return guide{}, errors.New("guide name cannot be empty")
	}
	g := guide{
		Name: name,
		Pois: []pointOfInterest{},
	}

	for _, opt := range opts {
		err := opt(&g)
		if err != nil {
			return guide{}, err
		}
	}
	return g, nil
}

// pointOfInterest represents a geo location in a map. Hence,
// the relationship is one-to-many, guide to points of Interests
// there is no guarantee that a poi is bounded within a maps coordinates. See IsBounded()
type pointOfInterest struct {
	Id          int64
	GuideID     int64
	Coordinate  coordinate
	Name        string
	Description string
}

// IsBounded determines if a pointOfInterest is bounded within guide.mapArea/coordinates
func (p pointOfInterest) IsBounded() bool {
	return false
}

type poiOption func(*pointOfInterest) error

func newPointOfInterest(name string, guideID int64, opts ...poiOption) (pointOfInterest, error) {
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

type guideForm struct {
	GuideId                                int64
	Name, Description, Latitude, Longitude string
	Errors                                 []string
}

type poiForm struct {
	GuideID                                int64
	GuideName                              string
	Name, Description, Latitude, Longitude string
	Errors                                 []string
}

type userForm struct {
	Username, Password, ConfirmPassword, Email string
	Errors                                     []string
}
