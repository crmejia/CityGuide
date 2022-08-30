package guide

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	store store
	*http.Server
}

func NewServer(address string, store store) (Server, error) {
	if address == "" {
		return Server{}, errors.New("server address cannot be empty")
	}
	if store == nil {
		return Server{}, errors.New("store cannot be nil")
	}

	server := Server{
		store: store,
		Server: &http.Server{
			Addr: address,
		},
	}

	server.Handler = server.routes()
	return server, nil
}

//go:embed templates
var fs embed.FS

func (s *Server) HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "templates/index.html", s.store.GetAllGuides())
	}
}

func (c *Server) HandleGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		if len(p) < 3 {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}

		guideID := p[2]
		id, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		g, err := c.store.GetGuide(id)
		if err != nil {
			http.Error(w, "Guide Not Found", http.StatusNotFound)
			return
		}

		render(w, r, "templates/guide.html", g)
	}
}

func (s *Server) HandleCreateGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			render(w, r, "templates/createGuide.html", nil)
			return
		}
		//http.MethodPost
		g := newGuideForm(w, r)
		if g == nil {
			return
		}
		gid, err := s.store.CreateGuide(*g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gURL := fmt.Sprintf("/guide/%d", gid)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
	}
}

func (s *Server) HandleCreatePoi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		if len(p) < 5 {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}
		guideID := p[4]
		gid, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "please provide valid guide id", http.StatusBadRequest)
		}
		g, err := s.store.GetGuide(gid)
		if err != nil {
			http.Error(w, "guide not found", http.StatusNotFound)
			return
		}
		poiForm := poiForm{
			GuideID:     gid,
			GuideName:   g.Name,
			Name:        r.PostFormValue("name"),
			Description: r.PostFormValue("description"),
			Latitude:    r.PostFormValue("latitude"),
			Longitude:   r.PostFormValue("longitude"),
		}
		if r.Method == http.MethodGet {
			render(w, r, "templates/createPoi.html", poiForm)
			return
		}
		//http.MethodPost
		_, err = NewPointOfInterest(poiForm.Name, gid, PoiWithValidStringCoordinates(poiForm.Latitude, poiForm.Longitude))
		if err != nil {
			poiForm.Errors = append(poiForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render(w, r, "templates/createPoi.html", poiForm)
			return
		}
		//TODO this should be a store operation func (s *store)CreatePoi(guideID, Poi{})(poiID, error)
		//g.Pois = append(*g.Pois, poi)
		gURL := fmt.Sprintf("/guide/%d", gid)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
	}
}
func (s *Server) Run() {
	log.Println("starting http server")
	err := s.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func ServerRun(address string) {
	store := memoryStore{
		Guides: map[int64]Guide{
			1: Guide{Id: 1, Name: "Nairobi", Coordinate: Coordinate{10, 10}},
			2: Guide{Id: 2, Name: "Fukuoka", Coordinate: Coordinate{11, 11}},
			3: Guide{Id: 3, Name: "Guia de restaurantes Roma, CDMX", Coordinate: Coordinate{12, 12}},
			4: Guide{Id: 4, Name: "Guia de Cuzco", Coordinate: Coordinate{13, 13}},
			5: Guide{Id: 5, Name: "San Cristobal de las Casas", Coordinate: Coordinate{Latitude: 16.7371, Longitude: -92.6375},
				Description: "Beatiful town in the mountains of the state of Chiapas.",
				Pois: &[]pointOfInterest{
					{Name: "Cafeología", Coordinate: Coordinate{16.737393, -92.635857}, Description: "Best Coffee in town. Maybe even the best coffee in the country."},
					{Name: "Centralita Coworking", Coordinate: Coordinate{16.739030, -92.635001}, Description: "Nice Coworking with a cool vibe."},
				}},
		},
	}
	store.nextGuideKey = 6
	s, err := NewServer(address, &store)
	if err != nil {
		log.Fatal(err)
	}
	s.Run()
}
func (s *Server) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/guide/", s.HandleGuide())
	router.HandleFunc("/guide/create/", s.HandleCreateGuide())
	router.HandleFunc("/guide/poi/create/", s.HandleCreatePoi())

	return router
}

func render(w http.ResponseWriter, r *http.Request, templateFile string, data any) {
	tmpl, err := template.ParseFS(fs, templateFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
