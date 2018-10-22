package setlxplayground

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
)

// RequestHandler answers all http requests
type RequestHandler struct {
	indexTemplate *template.Template
	bufpool       *bpool.BufferPool
	db            *CodeStorage
}

// NewRequestHandler creates a new handler with a own buffer pool
func NewRequestHandler(indexTemplate *template.Template, db *CodeStorage) *RequestHandler {
	return &RequestHandler{
		indexTemplate: indexTemplate,
		db:            db,
		bufpool:       bpool.NewBufferPool(64),
	}
}

// serves index page
func (h RequestHandler) index(w http.ResponseWriter, r *http.Request) {
	h.renderCodePage("print(\"Hello setlX\");", w, r)
}

// runs code and responds with console output
func (h RequestHandler) run(w http.ResponseWriter, r *http.Request) {
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
}

// stores code in database and responds with its id
func (h RequestHandler) share(w http.ResponseWriter, r *http.Request) {
	code, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Can not read code", http.StatusInternalServerError)
		return
	}
	id, err := h.db.SaveCode(code)
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
}

// serves code stored in database
func (h RequestHandler) code(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "invalid snippet id", http.StatusBadRequest)
		return
	}
	code, err := h.db.GetCode(id)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			http.Error(w, "snippet not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.renderCodePage(code, w, r)
}

// CodePageData is data for index page
type CodePageData struct {
	Code    string
	Embeded bool
	URL     string
}

// renders html page
func (h RequestHandler) renderCodePage(code string, w http.ResponseWriter, r *http.Request) {
	embeded, _ := strconv.ParseBool(r.URL.Query().Get("embeded"))
	query := r.URL.Query()
	query.Del("embeded")
	r.URL.RawQuery = query.Encode()

	buf := h.bufpool.Get()
	defer h.bufpool.Put(buf)

	err := h.indexTemplate.Execute(buf, CodePageData{
		Code:    code,
		Embeded: embeded,
		URL:     r.URL.String(),
	})
	if err != nil {
		log.Println("error executing template: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
