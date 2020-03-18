package main

func (s *Server) routes() {
	s.router.HandleFunc("/cardimage", s.handleGetCardImage()).Methods("GET")
	s.router.HandleFunc("/cardimagebatch", s.handlePostCardImageBatch()).Methods("POST")
	s.router.HandleFunc("/advertisementimages", s.handleGetAdvertisementImages()).Methods("GET")
	s.router.HandleFunc("/uploadimage", s.handleUploadImage()).Methods("POST")
	s.router.HandleFunc("/uploadimagebatch", s.handleUploadImageBatch()).Methods("POST")
}
