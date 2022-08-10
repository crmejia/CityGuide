package guide

import (
	"embed"
	_ "embed"
	"errors"
	"html/template"
	"log"
	"net/http"
)

type guide struct {
	Name string
}

type guideOption func(*guide) error

func WithValidCoordinates(latitude, longitude float64) guideOption {
	return func(g *guide) error {
		if latitude < -90 || latitude > 90 {
			return errors.New("latitude has to be in the -90°, 90° range")
		}
		if longitude < -180 || longitude > 180 {
			return errors.New("longitude has to be in the -90°, 90° range")
		}
		return nil
	}
}

//WithLocation
//WithLocationLookUp

//AreSpots outOfGuideBounds

func NewGuide(name string, opts ...guideOption) (guide, error) {
	if name == "" {
		return guide{}, errors.New("guide name cannot be empty")
	}
	g := guide{Name: name}
	for _, opt := range opts {
		err := opt(&g)
		if err != nil {
			return guide{}, err
		}
	}
	return g, nil
}

//go:embed templates
var fs embed.FS

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(fs, "templates/index.html")
	//template.ParseFiles("templates/index.html")
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

func ServerRun(address string) {
	http.HandleFunc("/", IndexHandler)
	log.Fatal(http.ListenAndServe(address, nil))
}
