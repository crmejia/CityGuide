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
	store store
	*http.Server
	output           io.Writer
	templateRegistry *templateRegistry
}

func NewServer(address string, store store, output io.Writer) (Server, error) {
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
		err := s.templateRegistry.render(w, indexTemplate, s.store.GetAllGuides())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		g, err := s.store.GetGuide(id)
		if err != nil {
			http.Error(w, "guide Not Found", http.StatusNotFound)
			return
		}
		g.Pois = s.store.GetAllPois(id)

		err = s.templateRegistry.render(w, guideTemplate, g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandleCreateGuideGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.templateRegistry.render(w, createGuideFormTemplate, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		g, err := s.store.CreateGuide(guideForm.Name, WithValidStringCoordinates(guideForm.Latitude, guideForm.Longitude), WithDescription(guideForm.Description))
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.render(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
		g, err := s.store.GetGuide(id)
		if err != nil {
			http.Error(w, "guide Not Found", http.StatusNotFound)
			return
		}
		if g.Id == 0 {
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
		err = s.templateRegistry.render(w, editGuideFormTemplate, guideForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		g, err := s.store.GetGuide(id)
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
			err := s.templateRegistry.render(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		g.Name = guideForm.Name
		g.Description = guideForm.Description
		g.Coordinate = coordinates

		err = s.store.UpdateGuide(&g)
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.render(w, createGuideFormTemplate, guideForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		gURL := fmt.Sprintf("/guide/%d", g.Id)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
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

		err = s.templateRegistry.render(w, createPoiFormTemplate, poiForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		_, err = s.store.CreatePoi(poiForm.Name, gid, PoiWithValidStringCoordinates(poiForm.Latitude, poiForm.Longitude), PoiWithDescription(poiForm.Description))
		if err != nil {
			poiForm.Errors = append(poiForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.render(w, createPoiFormTemplate, poiForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		gURL := fmt.Sprintf("/guide/%d", gid)
		http.Redirect(w, r, gURL, http.StatusSeeOther)
	}
}

func (s *Server) HandleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			err := s.templateRegistry.render(w, createUserFormTemplate, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		//http.MethodPost
		userForm := userForm{
			Username:        r.PostFormValue("username"),
			Password:        r.PostFormValue("password"),
			ConfirmPassword: r.PostFormValue("confirm-password"),
			Email:           r.PostFormValue("email"),
		}
		_, err := s.store.CreateUser(userForm.Username, userForm.Password, userForm.ConfirmPassword, userForm.Email)
		if err != nil {
			userForm.Errors = append(userForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.render(w, createUserFormTemplate, userForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
	store, err := OpenSQLiteStore(homeDir + "/city_guide.db")
	if err != nil {
		fmt.Fprintln(output, err)
		return
	}
	s, err := NewServer(address, &store, output)
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
	router.HandleFunc("/guide/{id}", s.HandleGuide())
	router.HandleFunc("/guide/{id}/edit", s.HandleEditGuideGet()).Methods(http.MethodGet)
	router.HandleFunc("/guide/{id}/edit", s.HandleEditGuidePost()).Methods(http.MethodPost)

	//POI *-> Guide
	router.HandleFunc("/guide/poi/create/{id}", s.HandleCreatePoiGet()).Methods(http.MethodGet)
	router.HandleFunc("/guide/poi/create/{id}", s.HandleCreatePoiPost()).Methods(http.MethodPost)
	router.HandleFunc("/user/signup", s.HandleCreateUser())
	router.HandleFunc("/", HandleIndex())
	return router
}

func templateRoutes() *templateRegistry {
	templates := map[string]*template.Template{}

	//todo iterate over template dir
	for _, templateName := range []string{indexTemplate, guideTemplate, createGuideFormTemplate, editGuideFormTemplate, createPoiFormTemplate, createUserFormTemplate} {
		templates[templateName] = template.Must(template.ParseFS(fs, templatesDir+templateName, templatesDir+baseTemplate))
	}

	return &templateRegistry{templates: templates}

}

//go:embed templates
var fs embed.FS

type templateRegistry struct {
	templates map[string]*template.Template
}

// w can be io.Writer or http.ResponseWriter. Keep it io to make sure we don't do http things here
func (t *templateRegistry) render(w io.Writer, templateFile string, data any) error {
	//tmpl, err := template.ParseFS(fs, templateFile)
	//err := tmpl.ExecuteTemplate(w,templateFile,data)
	tmpl, ok := t.templates[templateFile]
	if ok {
		return tmpl.ExecuteTemplate(w, baseTemplate, data)
	}
	err := errors.New("Template not found ->" + templateFile)
	return err
}

const (
	templatesDir            = "templates/"
	baseTemplate            = "base.html"
	indexTemplate           = "index.html"
	guideTemplate           = "guide.html"
	createGuideFormTemplate = "createGuideForm.html"
	editGuideFormTemplate   = "editGuideForm.html"
	createPoiFormTemplate   = "createPoiForm.html"
	createUserFormTemplate  = "createUserForm.html"
)
