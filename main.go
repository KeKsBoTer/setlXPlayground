package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"flag"
	tmpl "html/template"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/mux"
	"github.com/xlab/closer"
)

// CodePageData is data for index page
type CodePageData struct {
	Code    string
	Embeded bool
	URL     string
}

func main() {
	mode := flag.String("mode", "prod", "run mode, dev or prod")
	datbaseFolder := flag.String("database", "db", "folder of the database")
	port := flag.Int("port", 80, "port which the webserver listens on")
	flag.Parse()

	log.Printf("Starting setlX playground server in %s mode on port %d\n", *mode, *port)

	// load page html template
	template, err := tmpl.ParseFiles("www/index.html")
	if err != nil {
		log.Fatalln(err)
		return
	}

	// open database connection
	db, err := Open(*datbaseFolder)
	if err != nil {
		log.Fatalln(err)
		return
	}
	router := mux.NewRouter()
	router.StrictSlash(true)

	// index page handler
	router.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *mode == "dev" {
			template, err = tmpl.ParseFiles("www/index.html")
			if err != nil {
				log.Fatalln(err)
				return
			}
		}
		// check if page is embeded
		embeded, _ := strconv.ParseBool(r.URL.Query().Get("embeded"))
		err = template.Execute(w, CodePageData{
			Code:    "print(\"Hello setlX\");",
			Embeded: embeded,
		})
		if err != nil {
			log.Println("error executing template: ", err)
		}
	})

	// run code api
	// runs code and returns execution result
	router.Path("/run").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can not read code", http.StatusInternalServerError)
			return
		}
		output, err := run(code)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can not execute code", http.StatusInternalServerError)
			return
		}
		w.Write(output)
	})

	// share code api
	// takes code and returns code snippet id
	router.Path("/share").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can not read code", http.StatusInternalServerError)
			return
		}
		id, err := db.SaveCode(code)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can not share code", http.StatusInternalServerError)
			return
		}
		out, err := json.Marshal(struct {
			ID string `json:"id"`
		}{
			ID: id,
		})
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Write(out)
	})

	// shared code page
	router.Path("/c/{id:[a-zA-Z0-9]+}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "invalid snippet id", http.StatusBadRequest)
			return
		}
		code, err := db.GetCode(id)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				http.Error(w, "snippet not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		embeded, _ := strconv.ParseBool(r.URL.Query().Get("embeded"))
		query := r.URL.Query()
		query.Del("embeded")
		r.URL.RawQuery = query.Encode()
		err = template.Execute(w, CodePageData{
			Code:    code,
			Embeded: embeded,
			URL:     r.URL.String(),
		})
		if err != nil {
			log.Println("error executing template: ", err)
		}
	})

	// serve static files
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("www/static")))
	router.PathPrefix("/static/").Handler(fileHandler)

	server := &http.Server{Addr: ":" + strconv.Itoa(*port), Handler: router}

	closer.Bind(func() {
		if err := server.Close(); err != nil {
			log.Println("cannot close webserver:", err)
		} else {
			log.Println("closed webserver successfully")
		}
	})
	closer.Bind(func() {
		if err := db.Close(); err != nil {
			log.Println("Cannot close database connection:", err)
		} else {
			log.Println("closed database successfully")
		}
	})

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println("webserver failed:", err)
	}
}
