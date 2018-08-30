package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"flag"
	tmpl "html/template"

	"github.com/gorilla/mux"
)

var mode string

func main() {
	flag.StringVar(&mode, "mode", "prod", "run mode, dev or prod")
	datbaseFolder := *flag.String("database", "db/", "folder of the database")
	flag.Parse()

	template, err := tmpl.ParseFiles("www/index.html")
	if err != nil {
		log.Fatalln(err)
		return
	}

	db, err := Open(datbaseFolder)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if mode == "dev" {
			template, err = tmpl.ParseFiles("www/index.html")
			if err != nil {
				log.Fatalln(err)
				return
			}
		}
		template.Execute(w, nil)
	})

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

	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("tmp", os.ModeTemporary); err != nil {
			log.Fatalln(err)
			return
		}
	}
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("www/static")))
	router.PathPrefix("/static/").Handler(fileHandler)
	http.ListenAndServe(":8080", router)
}
