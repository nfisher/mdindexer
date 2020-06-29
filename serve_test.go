package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_routes_ok(t *testing.T) {
	cases := map[string]struct {
		method      string
		path        string
		contentType string
	}{
		"search":  {http.MethodGet, "/search?q=development", ApplicationJson},
		"file":    {http.MethodGet, "/files/hello.html", TextHtml},
		"root":    {http.MethodGet, "/", TextHtml},
		"main.js": {http.MethodGet, "/main.js", ApplicationJs},
	}
	index := New(10)
	index.Update(&Document{Name: "index.md", WordCount: map[string]int{"development": 1}})
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost%s", tc.path)
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Errorf("NewRequest(%s, %s, ...) error=%v, want nil", tc.method, url, err)
			}

			mux := BuildRoutes("./testdata", index)
			mux.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Errorf("w.Code=%d, want 200 OK", w.Code)
			}
			actual := w.Header().Get(HeaderContentType)
			if actual != tc.contentType {
				t.Errorf("Content-type=<%s>, want <%s> - %s", actual, tc.contentType, w.Body.String())
			}
		})
	}
}
