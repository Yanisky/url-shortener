package api

// Route Attaches handlers to routes
func (s *Server) Route(handler URLShortnerHttpHandler) {
	s.Router.HandleFunc("/{urlHash}", handler.Redirect).Methods("GET")
	s.Router.HandleFunc("/api/v1/urls", handler.CreateURL).Methods("POST")
	s.Router.HandleFunc("/api/v1/urls/{urlHash}/views", handler.ViewUrlStats).Methods("GET")
}
