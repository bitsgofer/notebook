package blog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bitsgofer/notebook/internal/middlewares/redirect"
)

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
			body, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if want, got := tc.expectedBody, string(body); want != got {
				t.Errorf("wrote wrong body,\n  want= %q\n   got= %q", want, got)
			}
		})
	}
}

func TestBlogHandlerServeError(t *testing.T) {
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
			handler := blogHandler("testdata")

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

func TestServeErrPage(t *testing.T) {
	var testCases = []struct {
		status         int
		expectedStatus int
		expectedBody   string
	}{
		{
			status:         http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "<html><h1>Not found</h1><p>Sorry, but our princess is in another castle</p></html>",
		},
		{
			status:         http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "<html><h1>Bad request</h1><p>Sorry, this we can't serve this</p></html>",
		},
		{
			status:         http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "<html><h1>Internal server error</h1><p>Sorry, something went wrong</p></html>",
		},
		{
			status:         600,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "<html><h1>Internal server error</h1><p>Sorry, something went wrong</p></html>",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			serveErrPage(w, req, tc.status)
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

var (
	exampleDomains = []string{
		"example.com",
		"www.example.com",
	}
)

func TestHTTPRedirectHandler(t *testing.T) {
	srv, _ := New("testdata", "admin@example.com", exampleDomains)
	path := "example.com/somewhere"
	req := httptest.NewRequest(http.MethodGet, "http://"+path, nil)
	w := httptest.NewRecorder()

	srv.HTTPRedirectHandler().ServeHTTP(w, req)
	resp := w.Result()
	if want, got := http.StatusFound, resp.StatusCode; want != got {
		t.Errorf("wrote wrong HTTP status, want= %v, got= %v", want, got)
	}
	if want, got := "https://"+path, resp.Header.Get("Location"); want != got {
		t.Errorf("wrote wrong Location header, want= %v, got= %v", want, got)
	}
}

func TestBlogHandler(t *testing.T) {
	serveFile = func(w http.ResponseWriter, r *http.Request, fname string) {
		http.ServeFile(w, r, fname)
	}

	srv, _ := New("testdata", "admin@example.com", exampleDomains)
	path := "example.com/sample"
	req := httptest.NewRequest(http.MethodGet, "https://"+path, nil)
	w := httptest.NewRecorder()
	expectedBody := "<html><head></head><body><h1>sample</h1><p>body</p></body></html>"

	srv.BlogHandler().ServeHTTP(w, req)
	resp := w.Result()
	if want, got := http.StatusOK, resp.StatusCode; want != got {
		t.Errorf("wrote wrong HTTP status, want= %v, got= %v", want, got)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if want, got := expectedBody, string(body); want != got {
		t.Errorf("wrote wrong body,\n  want= %q\n   got= %q", want, got)
	}
}

func TestBlogHandlerForward(t *testing.T) {
	serveFile = func(w http.ResponseWriter, r *http.Request, fname string) {
		http.ServeFile(w, r, fname)
	}

	from, _ := url.Parse("https://subdomain.example.com/")
	to, _ := url.Parse("https://to.forward.domain/path?query=val")

	srv, err := New("testdata", "admin@example.com",
		append(exampleDomains, "subdomain.example.com"),
		Redirect(redirect.Redirections{
			redirect.Redirection{FromURL: *from, ToURL: *to},
		}),
	)
	if err != nil {
		t.Fatalf("err= %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://subdomain.example.com/", nil)
	w := httptest.NewRecorder()

	srv.BlogHandler().ServeHTTP(w, req)

	resp := w.Result()
	if want, got := http.StatusMovedPermanently, resp.StatusCode; want != got {
		t.Errorf("wrote wrong HTTP status, want= %v, got= %v", want, got)
	}
	if want, got := to.String(), resp.Header.Get("Location"); want != got {
		t.Errorf("forwarded to wrong URL\n  want= %q\n   got= %q", want, got)
	}
}
