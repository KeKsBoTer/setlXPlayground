package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"flag"
	tmpl "html/template"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/mux"
)

var mode string

func main() {
	flag.StringVar(&mode, "mode", "prod", "run mode, dev or prod")
	datbaseFolder := *flag.String("database", "db", "folder of the database")
	port := *flag.Int("port", 80, "port which the webserver listens on")
	flag.Parse()

	log.Printf("Starting setlX playground server in %s mode on port %d\n", mode, port)

	// load page html template
	template, err := tmpl.ParseFiles("www/index.html")
	if err != nil {
		log.Fatalln(err)
		return
	}

	// open database connection
	db, err := Open(datbaseFolder)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()

	router := mux.NewRouter()
	router.StrictSlash(true)

	// index page handler
	router.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mode == "dev" {
			template, err = tmpl.ParseFiles("www/index.html")
			if err != nil {
				log.Fatalln(err)
				return
			}
		}
		template.Execute(w, "print(\"Hello setlX\");")
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
		template.Execute(w, code)
	})

	// create dir for code files
	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("tmp", os.ModeTemporary); err != nil {
			log.Fatalln(err)
			return
		}
	}

	// serve static files
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("www/static")))
	router.PathPrefix("/static/").Handler(fileHandler)

	http.ListenAndServe(":"+strconv.Itoa(port), router)
}
