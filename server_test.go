package lazypress

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

func TestShouldDownloadPDFWhenNoOutputParam(t *testing.T) {
	html := `<html><body>Hello World</body></html>`
	req, err := http.NewRequest("POST", "/convert", strings.NewReader(html))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Content-Length", fmt.Sprint((len(html))))
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.Header.Get("Content-Type") != "application/pdf" {
		t.Errorf("Expected Content-Type to be application/pdf, got %s", result.Header.Get("Content-Type"))
	}
	// TODO: add some more robust checks
}

func TestShouldReturnErrorWhenContentTypeIsNotAllowed(t *testing.T) {
	html := `<html><body>Hello World</body></html>`
	req, err := http.NewRequest("POST", "/convert", strings.NewReader(html))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprint((len(html))))
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be 400, got %d", result.StatusCode)
	}
}

func TestShouldReturnErrorWhenContentLenghtIsZero(t *testing.T) {
	html := `<html><body>Hello World</body></html>`
	req, err := http.NewRequest("POST", "/convert", strings.NewReader(html))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Content-Length", "0")
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be 400, got %d", result.StatusCode)
	}
}

func TestShouldReturnErrorWhenRequestBodyIsEmpty(t *testing.T) {
	req, err := http.NewRequest("POST", "/convert", strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "text/html")
	// we don't want to test the content length
	// validation (we did this before).
	// we want to just test the actual size of
	// the req's body.
	// therefore we set the content-length to
	// a non-zero value.
	req.Header.Set("Content-Length", "42")
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be 400, got %d", result.StatusCode)
	}
}

func TestShouldReturnErrorIfSanitizationIsOnAndContentIsAllScript(t *testing.T) {
	html := `<script>alert("Hello World")</script>`
	req, err := http.NewRequest("POST", "/convert?sanitize=true", strings.NewReader(html))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Content-Length", fmt.Sprint((len(html))))
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be 400, got %d", result.StatusCode)
	}
}

func TestShouldNotReturnErrorIfSanitizationIsOffAndContentIsAllScript(t *testing.T) {
	html := `<script>alert("Hello World")</script>`
	req, err := http.NewRequest("POST", "/convert", strings.NewReader(html))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Content-Length", fmt.Sprint((len(html))))
	w := httptest.NewRecorder()
	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		log.Fatalln(dirError)
	}
	chromePath := path.Join(dir, "chrome-linux", "chrome")
	convertHTMLServerHandler(chromePath)(w, req)
	result := w.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code to be 200, got %d", result.StatusCode)
	}
	if result.Header.Get("Content-Type") != "application/pdf" {
		t.Errorf("Expected Content-Type to be application/pdf, got %s", result.Header.Get("Content-Type"))
	}
}
