package guide

import (
	"embed"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	store Storage
	*http.Server
	output           io.Writer
	templateRegistry *templateRegistry
}

func NewServer(address string, store Storage, output io.Writer) (Server, error) {
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
		output: output,
	}

	server.templateRegistry = templateRoutes()
	server.Handler = server.Routes()
	return server, nil
}

func HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "guides", http.StatusFound)
	}
}

func (s *Server) HandleGuides() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		terms := r.URL.Query().Get("q")
		guides, err := s.store.Search(terms)
		if err != nil || len(guides) == 0 {
			http.Error(w, "no guide found", http.StatusNotFound)
			return
		}

		if r.Header.Get("HX-Trigger") == "search" {
			err = s.templateRegistry.renderPartial(w, rowsTemplate, guides)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		err = s.templateRegistry.renderPage(w, indexTemplate, guides)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandleGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		g, err := s.store.GetGuidebyID(id)
		if err != nil {
			http.Error(w, "guide Not Found", http.StatusInternalServerError)
			return
		}
		if g == nil {
			http.Error(w, "guide Not Found", http.StatusNotFound)
			return

		}
		g.Pois = s.store.GetAllPois(id)

		err = s.templateRegistry.renderPage(w, guideTemplate, g)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandleCreateGuideGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.templateRegistry.renderPage(w, createGuideFormTemplate, nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (s *Server) HandleCreateGuidePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideForm := guideForm{
			Name:        r.PostFormValue("name"),
			Description: r.PostFormValue("description"),
			Latitude:    r.PostFormValue("latitude"),
			Longitude:   r.PostFormValue("longitude"),
			Errors:      []string{},
		}
		g, err := NewGuide(guideForm.Name, WithValidStringCoordinates(guideForm.Latitude, guideForm.Longitude), WithDescription(guideForm.Description))
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		err = s.store.CreateGuide(&g)
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		gURL := fmt.Sprintf("/guide/%d", g.Id)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
	}

}

func (s *Server) HandleEditGuideGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		g, err := s.store.GetGuidebyID(id)
		if err != nil {
			http.Error(w, "db error", http.StatusNotFound)
			return
		}
		if g == nil {
			http.Error(w, "guide not found", http.StatusNotFound)
			return
		}

		guideForm := guideForm{
			GuideId:     g.Id,
			Name:        g.Name,
			Description: g.Description,
			Latitude:    fmt.Sprintf("%f", g.Coordinate.Latitude),
			Longitude:   fmt.Sprintf("%f", g.Coordinate.Longitude),
			Errors:      []string{},
		}
		err = s.templateRegistry.renderPage(w, editGuideFormTemplate, guideForm)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (s *Server) HandleEditGuidePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		g, err := s.store.GetGuidebyID(id)
		if err != nil {
			http.Error(w, "guide Not Found", http.StatusNotFound)
			return
		}
		if g.Id == 0 {
			http.Error(w, "guide not found", http.StatusNotFound)
			return
		}

		guideForm := guideForm{
			Name:        r.PostFormValue("name"),
			Description: r.PostFormValue("description"),
			Latitude:    r.PostFormValue("latitude"),
			Longitude:   r.PostFormValue("longitude"),
			Errors:      []string{},
		}

		coordinates, err := parseCoordinates(guideForm.Latitude, guideForm.Longitude)
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		g.Name = guideForm.Name
		g.Description = guideForm.Description
		g.Coordinate = coordinates

		err = s.store.UpdateGuide(g)
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		gURL := fmt.Sprintf("/guide/%d", g.Id)
		http.Redirect(w, r, gURL, http.StatusOK)
	}

}

func (s *Server) HandleDeleteGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		err = s.store.DeleteGuide(id)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if mux.Vars(r)["HX-Trigger"] == "delete-btn" {
			http.Redirect(w, r, "/guides", http.StatusSeeOther)
		}
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (s *Server) HandleCreatePoiGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}

		gid, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "please provide valid guide Id", http.StatusBadRequest)
		}

		g, err := s.store.GetGuidebyID(gid)
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

		err = s.templateRegistry.renderPage(w, createPoiFormTemplate, poiForm)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (s *Server) HandleCreatePoiPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideID := mux.Vars(r)["id"]
		if guideID == "" {
			http.Error(w, "no guideid provided", http.StatusBadRequest)
			return
		}

		gid, err := strconv.ParseInt(guideID, 10, 64)
		if err != nil {
			http.Error(w, "please provide valid guide Id", http.StatusBadRequest)
		}

		g, err := s.store.GetGuidebyID(gid)
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
		poi, err := NewPointOfInterest(poiForm.Name, gid, PoiWithValidStringCoordinates(poiForm.Latitude, poiForm.Longitude), PoiWithDescription(poiForm.Description))
		if err != nil {
			poiForm.Errors = append(poiForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createPoiFormTemplate, poiForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		err = s.store.CreatePoi(&poi)
		if err != nil {
			poiForm.Errors = append(poiForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.renderPage(w, createPoiFormTemplate, poiForm)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		gURL := fmt.Sprintf("/guide/%d", gid)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
	}
}

func (s *Server) HandleDeletePoi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guideIDString := mux.Vars(r)["guideID"]
		if guideIDString == "" {
			http.Error(w, "no guide ID provided", http.StatusBadRequest)
			return
		}
		poiIDString := mux.Vars(r)["id"]
		if poiIDString == "" {
			http.Error(w, "no poi ID provided", http.StatusBadRequest)
			return
		}
		guideID, err := strconv.ParseInt(guideIDString, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse guide ID", http.StatusBadRequest)
			return
		}
		poiID, err := strconv.ParseInt(poiIDString, 10, 64)
		if err != nil {
			http.Error(w, "not able to parse poi ID", http.StatusBadRequest)
			return
		}

		err = s.store.DeletePoi(guideID, poiID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if mux.Vars(r)["HX-Trigger"] == "delete-btn" {
			http.Redirect(w, r, "/guides", http.StatusSeeOther)
		}
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (s *Server) Run() {
	fmt.Fprintln(s.output, "starting http server")
	err := s.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Fprintln(s.output, err)
		return
	}
}

func RunServer(output io.Writer) {
	address := os.Getenv("ADDRESS")
	if address == "" {
		fmt.Fprintln(output, "no address provided, defaulting to :8080")
		address = ":8080"
	}
	homeDir, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(output, err)
		return
	}
	storage, err := OpenSQLiteStorage(homeDir + "/city_guide.db")
	if err != nil {
		fmt.Fprintln(output, err)
		return
	}
	s, err := NewServer(address, storage, output)
	if err != nil {
		fmt.Fprintln(output, err)
		return
	}
	s.Run()
}
func (s *Server) Routes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/guides", s.HandleGuides())
	router.HandleFunc("/guide/create", s.HandleCreateGuideGet()).Methods(http.MethodGet)
	router.HandleFunc("/guide/create", s.HandleCreateGuidePost()).Methods(http.MethodPost)
	router.HandleFunc("/guide/{id}", s.HandleDeleteGuide()).Methods(http.MethodDelete)
	router.HandleFunc("/guide/{id}", s.HandleGuide())
	router.HandleFunc("/guide/{id}/edit", s.HandleEditGuideGet()).Methods(http.MethodGet)
	router.HandleFunc("/guide/{id}/edit", s.HandleEditGuidePost()).Methods(http.MethodPost)

	//POI *-> guide
	router.HandleFunc("/guide/poi/create/{id}", s.HandleCreatePoiGet()).Methods(http.MethodGet)
	router.HandleFunc("/guide/poi/create/{id}", s.HandleCreatePoiPost()).Methods(http.MethodPost)
	router.HandleFunc("/guide/{guideID}/poi/{id}", s.HandleDeletePoi()).Methods(http.MethodDelete)
	router.HandleFunc("/", HandleIndex())
	return router
}

func templateRoutes() *templateRegistry {
	pageTemplates := map[string]*template.Template{}
	partialTemplates := map[string]*template.Template{}

	//todo iterate over template dir
	for _, templateName := range []string{indexTemplate, guideTemplate, createGuideFormTemplate, editGuideFormTemplate, createPoiFormTemplate} {
		pageTemplates[templateName] = template.Must(template.ParseFS(fs, templatesDir+templateName, templatesDir+baseTemplate, templatesDir+rowsTemplate))
	}
	for _, templateName := range []string{rowsTemplate} {
		partialTemplates[templateName] = template.Must(template.ParseFS(fs, templatesDir+templateName))
	}

	return &templateRegistry{
		pageTemplates:    pageTemplates,
		partialTemplates: partialTemplates,
	}

}

//go:embed templates
var fs embed.FS

type templateRegistry struct {
	pageTemplates    map[string]*template.Template
	partialTemplates map[string]*template.Template
}

// w can be io.Writer or http.ResponseWriter. Keep it io to make sure we don't do http things here
func (t *templateRegistry) renderPage(w io.Writer, templateFile string, data any) error {
	tmpl, ok := t.pageTemplates[templateFile]
	if ok {
		//todo in theory tmpl.Execute should be enough as template already created with base
		return tmpl.ExecuteTemplate(w, baseTemplate, data)

	}
	return errors.New("Template not found ->" + templateFile)
}

func (t *templateRegistry) renderPartial(w io.Writer, templateFile string, data any) error {
	tmpl, ok := t.partialTemplates[templateFile]
	if ok {
		return tmpl.Execute(w, data)
		//return tmpl.ExecuteTemplate(w, rowsTemplate, data)
	}
	return errors.New("Template not found ->" + templateFile)
}

const (
	templatesDir            = "templates/"
	baseTemplate            = "base.html"
	indexTemplate           = "index.html"
	rowsTemplate            = "rows.html"
	guideTemplate           = "guide.html"
	createGuideFormTemplate = "createGuideForm.html"
	editGuideFormTemplate   = "editGuideForm.html"
	createPoiFormTemplate   = "createPoiForm.html"
)
