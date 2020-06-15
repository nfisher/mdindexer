package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_routes_ok(t *testing.T) {
	cases := map[string]struct {
		method      string
		path        string
		body        io.Reader
		contentType string
	}{
		"search":  {http.MethodGet, "/search?q=development", nil, ApplicationJson},
		"file":    {http.MethodGet, "/files/hello.html", nil, TextHtml},
		"root":    {http.MethodGet, "/", nil, TextHtml},
		"main.js": {http.MethodGet, "/main.js", nil, ApplicationJs},
	}
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost%s", tc.path)
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tc.method, url, tc.body)
			if err != nil {
				t.Errorf("NewRequest(%s, %s, ...) error=%v, want nil", tc.method, url, err)
			}

			index := New(10)
			mux := BuildRoutes("./testdata", index)
			mux.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Errorf("w.Code=%d, want 200 OK", w.Code)
			}
			actual := w.Header().Get(HeaderContentType)
			if actual != tc.contentType {
				t.Errorf("Content-type=<%s>, want <%s>", actual, tc.contentType)
			}
		})
	}
}
