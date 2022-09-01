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
			http.Error(w, "guide Not Found", http.StatusNotFound)
			return
		}
		g.Pois = c.store.GetAllPois(id)

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
		guideForm := guideForm{
			Name:        r.PostFormValue("name"),
			Description: r.PostFormValue("description"),
			Latitude:    r.PostFormValue("latitude"),
			Longitude:   r.PostFormValue("longitude"),
		}
		g, err := s.store.CreateGuide(guideForm.Name, GuideWithValidStringCoordinates(guideForm.Latitude, guideForm.Longitude))
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render(w, r, "templates/createGuide.html", guideForm)
			return
		}
		gURL := fmt.Sprintf("/guide/%d", g.Id)
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

		_, err = s.store.CreatePoi(poiForm.Name, gid, PoiWithValidStringCoordinates(poiForm.Latitude, poiForm.Longitude), PoiWithDescription(poiForm.Description))
		if err != nil {
			poiForm.Errors = append(poiForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render(w, r, "templates/createPoi.html", poiForm)
			return
		}

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
	store, err := OpenSQLiteStore("city_guide.db")
	if err != nil {
		log.Fatal(err)
	}
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
