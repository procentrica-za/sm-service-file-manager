package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var conf Config

func init() {
	conf = CreateConfig()
}

func CreateConfig() Config {
	conf := Config{
		ListenServePort: os.Getenv("FILE_MANAGER_PORT"),
		ResourcesPath:   os.Getenv("RESOURCES_PATH"),
		CRUDHost:        os.Getenv("CRUD_HOST"),
		CRUDPort:        os.Getenv("CRUD_PORT"),
	}
	return conf
}

func main() {
	server := Server{
		router: mux.NewRouter(),
	}

	// Setup Routes for the server
	server.routes()
	handler := removeTrailingSlash(server.router)

	fmt.Printf("Starting service on port -->  " + conf.ListenServePort + " .... \n")
	log.Fatal(http.ListenAndServe(":"+conf.ListenServePort, handler))
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}
