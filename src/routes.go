package main

func (s *Server) routes() {
	s.router.HandleFunc("/cardimage", s.handleGetCardImage()).Methods("GET")
	s.router.HandleFunc("/cardimagebatch", s.handlePostCardImageBatch()).Methods("POST")
}
