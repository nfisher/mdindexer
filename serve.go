package main

import (
	"encoding/json"
	"github.com/rakyll/statik/fs"
	"log"
	"net/http"
	"path/filepath"

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
	mux := http.NewServeMux()

	files, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(files))

	for _, p := range paths {
		prefix := filepath.Join("/files", p)
		mux.Handle(prefix + "/", http.StripPrefix(prefix, http.FileServer(http.Dir(p))))
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
