package blogserver

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestNewServer(t *testing.T) {
// 	srv, err := New()
// 	if err != nil {
// 	}
//
// 	// check srv
// }
//
// func TestHTTPHandler(t *testing.T) {
// 	// call srv.HTTPHandler -> func(http.ResponseWriter, *http.Request)
// 	// it should handle ACME + redirects the rest to HTTPS
// }
//
// func TestHTTPHandler(t *testing.T) {
// 	// create srv with a custom handler
// 	// call srv.HTTPSHandler -> same as the custom handler
// }
//
// func TestReload(t *testing.T) {
// 	// modify root dir
// 	// call srv.Reload()
// 	// check that newly generated is okay
// }

func TestBlogHandlerServeFile(t *testing.T) {
	rootDir := "testdata"
	var testCases = []struct {
		name         string
		path         string
		expectedFile string
		expectedBody string
	}{
		{
			name:         "filename only",
			path:         "/sample",
			expectedFile: rootDir + "/sample.html",
			expectedBody: "<html><head></head><body><h1>sample</h1><p>body</p></body></html>",
		},
		{
			name:         ".html",
			path:         "/sample.html",
			expectedFile: rootDir + "/sample.html",
			expectedBody: "<html><head></head><body><h1>sample</h1><p>body</p></body></html>",
		},
		{
			name:         ".css",
			path:         "/style.css",
			expectedFile: rootDir + "/style.css",
			expectedBody: "@charset \"UTF-8\"",
		},
		{
			name:         ".js",
			path:         "/script.js",
			expectedFile: rootDir + "/script.js",
			expectedBody: "alert(\"hello, world!\");",
		},
		{
			name:         ".ico",
			path:         "/favicon.ico",
			expectedFile: rootDir + "/favicon.ico",
			expectedBody: string([]byte{0x00, 0x99, 0xaa, 0xFF}),
		},
		{
			name:         "root",
			path:         "/",
			expectedFile: rootDir + "/index.html",
			expectedBody: "<p>index</p>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			served := make(chan string)
			serveFile = func(w http.ResponseWriter, r *http.Request, fname string) {
				http.ServeFile(w, r, fname)
				served <- fname
			}
			handler := blogHandler(rootDir)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			go handler(w, req)
			if want, got := tc.expectedFile, <-served; want != got {
				t.Errorf("served wrong file, want= %v, got= %v", want, got)
			}
			resp := w.Result()
			if want, got := http.StatusOK, resp.StatusCode; want != got {
				t.Errorf("wrote wrong HTTP status, want= %v, got= %v", want, got)
			}
			body, err := ioutil.ReadAll(resp.Body)
			t.Logf("err= %v", err)
			defer resp.Body.Close()
			if want, got := tc.expectedBody, string(body); want != got {
				t.Errorf("wrote wrong body,\n  want= %q\n   got= %q", want, got)
			}
		})
	}
}

func TestBlogHandlerServeError(t *testing.T) {
	rootDir := "testdata"
	var testCases = []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST",
			method:         http.MethodPost,
			path:           "/sample",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   string(defaultResponses[http.StatusBadRequest]),
		},
		{
			name:           "non-existing file",
			method:         http.MethodGet,
			path:           "/non-existing",
			expectedStatus: http.StatusNotFound,
			expectedBody:   string(defaultResponses[http.StatusNotFound]),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serveFile = func(w http.ResponseWriter, r *http.Request, fname string) {} // NOP
			handler := blogHandler(rootDir)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()
			if want, got := tc.expectedStatus, resp.StatusCode; want != got {
				t.Errorf("wrote wrong HTTP status, want= %v, got= %v", want, got)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			if want, got := tc.expectedBody, string(body); want != got {
				t.Errorf("wrote wrong body,\n  want= %q\n   got= %q", want, got)
			}
		})
	}
}
