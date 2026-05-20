package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	host string
	port int
	mux  *http.ServeMux
}

func New(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) Setup() error {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	s.mux.HandleFunc("/", s.handleIndex)

	return nil
}

func (s *Server) RegisterAPI(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("INFO [server] Starting on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
		return
	}
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}
