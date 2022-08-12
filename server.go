package guide

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
)

type Server struct {
	store *memoryStore
	*http.Server
}

func NewServer(address string, store *memoryStore) (Server, error) {
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
		tmpl, err := template.ParseFS(fs, "templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, s.store.GetAllGuides())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// func (c * Server)
//func (c *Server) GuideHandler(w http.ResponseWriter, r *http.Request) {
//	guideID := r.FormValue("guideid")
//	if guideID == "" {
//		http.Error(w, "no guideid provided", http.StatusBadRequest)
//		return
//	}
//	Id, err := strconv.Atoi(guideID)
//	if err != nil {
//		http.Error(w, "unparsable Guide Id", http.StatusBadRequest)
//	}
//	//g, ok := c.Guides[Id]
//	if !ok {
//		http.Error(w, "Guide not found", http.StatusNotFound)
//	}
//
//	tmpl, err := template.ParseFS(fs, "templates/html")
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	err = tmpl.Execute(w, g)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//}

func (s *Server) Run() {
	log.Println("starting http server")
	err := s.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func ServerRun(address string) {
	store := memoryStore{
		Guides: map[int]Guide{
			1: Guide{Name: "Nairobi", Coordinate: Coordinate{10, 10}},
			2: Guide{Name: "Fukuoka", Coordinate: Coordinate{11, 11}},
			3: Guide{Name: "Guia de restaurantes Roma, CDMX", Coordinate: Coordinate{12, 12}},
			4: Guide{Name: "Guia de Cuzco", Coordinate: Coordinate{13, 13}},
		}}
	s, err := NewServer(address, &store)
	if err != nil {
		log.Fatal(err)
	}
	s.Run()
}
func (s *Server) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", s.HandleIndex())

	return router
}
