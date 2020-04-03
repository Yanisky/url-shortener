package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func NewGorillaHttpServer() Server {
	return Server{
		Router: mux.NewRouter(),
	}
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Router)
}
