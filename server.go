package guide

import (
	"embed"
	"errors"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	server.Handler = server.routes()
	return server, nil
}

func (s *Server) HandleIndex() http.HandlerFunc {
	//tmpls := []string{"templates/base.html", "templates/index.html"}
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.templateRegistry.render(w, indexTemplate, s.store.GetAllGuides())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) HandleGuide() http.HandlerFunc {
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

func (s *Server) HandleCreateGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			err := s.templateRegistry.render(w, createGuideFormTemplate, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		//http.MethodPost
		guideForm := guideForm{
			Name:        r.PostFormValue("name"),
			Description: r.PostFormValue("description"),
			Latitude:    r.PostFormValue("latitude"),
			Longitude:   r.PostFormValue("longitude"),
		}
		g, err := s.store.CreateGuide(guideForm.Name, GuideWithValidStringCoordinates(guideForm.Latitude, guideForm.Longitude), GuideWithDescription(guideForm.Description))
		if err != nil {
			guideForm.Errors = append(guideForm.Errors, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			err := s.templateRegistry.render(w, createGuideFormTemplate, guideForm.Errors)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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
		if r.Method == http.MethodGet {
			err := s.templateRegistry.render(w, createPoiFormTemplate, poiForm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
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
func (s *Server) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/guide/", s.HandleGuide())
	router.HandleFunc("/guide/create/", s.HandleCreateGuide())
	router.HandleFunc("/guide/poi/create/", s.HandleCreatePoi())
	router.HandleFunc("/user/signup/", s.HandleCreateUser())

	return router
}

func templateRoutes() *templateRegistry {
	templates := map[string]*template.Template{}

	for _, templateName := range []string{indexTemplate, guideTemplate, createGuideFormTemplate, createPoiFormTemplate, createUserFormTemplate} {
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
	createUserFormTemplate  = "createUserForm.html"
	createGuideFormTemplate = "createGuideForm.html"
	createPoiFormTemplate   = "createPoiForm.html"
)
