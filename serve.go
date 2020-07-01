package main

import (
	"encoding/json"
	"fmt"
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

func BuildRoutes(paths []string, index *Index) *http.ServeMux {
	files, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(files))
	for _, p := range paths {
		prefix := fmt.Sprintf("/files/%s/", p)
		mux.Handle(prefix, http.StripPrefix("/files/", http.FileServer(http.Dir("."))))
	}
	mux.HandleFunc("/search", SearchIndex(index))
	return mux
}

type SearchResponse struct {
	Docs ScoreList
}

func SearchIndex(index *Index) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		needle := r.URL.Query().Get("q")
		docs := Search(needle, index)
		w.Header().Set(HeaderContentType, ApplicationJson)
		err := json.NewEncoder(w).Encode(&SearchResponse{Docs: docs})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
