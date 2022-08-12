package guide

import (
	"embed"
	_ "embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Controller struct {
	Guides map[int]Guide
}
type Guide struct {
	Name       string
	Coordinate Coordinate
	pois       map[Coordinate]pointOfInterest
}

type pointOfInterest struct {
	Coordinate Coordinate
	Name       string
}

type Coordinate struct {
	Latitude, Longitude float64
}

func NewCoordinate(latitude, longitude float64) (Coordinate, error) {
	if latitude < -90 || latitude > 90 {
		return Coordinate{}, errors.New("Latitude has to be in the -90°, 90° range")
	}
	if longitude < -180 || longitude > 180 {
		return Coordinate{}, errors.New("Longitude has to be in the -90°, 90° range")
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

//WithLocation
//WithLocationLookUp

//AreSpots outOfGuideBounds

func NewGuide(name string, opts ...guideOption) (Guide, error) {
	if name == "" {
		return Guide{}, errors.New("Guide Name cannot be empty")
	}
	g := Guide{
		Name: name,
		pois: map[Coordinate]pointOfInterest{},
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
	g.pois[poi.Coordinate] = poi
	return nil
}

//go:embed templates
var fs embed.FS

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(fs, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GuideHandler(w http.ResponseWriter, r *http.Request) {
	guideID := r.FormValue("guideid")
	if guideID == "" {
		http.Error(w, "no guideid provided", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(guideID)
	if err != nil {
		http.Error(w, "unparsable Guide id", http.StatusBadRequest)
	}
	g, ok := c.Guides[id]
	if !ok {
		http.Error(w, "Guide not found", http.StatusNotFound)
	}

	tmpl, err := template.ParseFS(fs, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ServerRun(address string) {
	http.HandleFunc("/", IndexHandler)
	log.Fatal(http.ListenAndServe(address, nil))
}
