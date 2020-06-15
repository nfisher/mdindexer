package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rakyll/statik/fs"

	_ "github.com/nfisher/mdindexer/statik"
)

//go:generate statik -m -src=./_tpl

const (
	HeaderContentType = `Content-type`
	TextHtml          = `text/html; charset=utf-8`
	ApplicationJson   = `application/json`
	ApplicationJs     = `application/javascript`
)

func BuildRoutes(path string, index *Index) *http.ServeMux {
	files, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(files))
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(path))))
	mux.HandleFunc("/search", SearchIndex(index))
	return mux
}

type SearchResponse struct {
	Docs DocList
}

func SearchIndex(index *Index) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		enc := json.NewEncoder(w)
		var resp SearchResponse
		needle := r.URL.Query().Get("q")
		var docs = DocList{}
		if needle != "" {
			matches := ExactMatch(needle, index)
			for k := range matches {
				docs = append(docs, k)
			}
		}
		w.Header().Set(HeaderContentType, ApplicationJson)
		resp.Docs = docs
		err = enc.Encode(&resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
