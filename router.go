package setlxplayground

import (
	"net/http"

	"github.com/gorilla/mux"
)

// CreateRouter maps RequestHandler functions to REST API
func CreateRouter(handler *RequestHandler) *mux.Router {

	router := mux.NewRouter()
	router.StrictSlash(true)

	// index page handler
	router.Path("/").Methods("GET").HandlerFunc(handler.index)

	// run code api
	// runs code and returns execution result
	router.Path("/run").Methods("POST").HandlerFunc(handler.run)

	// share code api
	// takes code and returns code snippet id
	router.Path("/share").Methods("POST").HandlerFunc(handler.share)

	// shared code page
	router.Path("/c/{id:[a-zA-Z0-9]+}").Methods("GET").HandlerFunc(handler.code)

	// serve static files
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("www/static")))
	router.PathPrefix("/static/").Handler(fileHandler)

	return router
}
