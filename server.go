package lazypress

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/chromedp/cdproto/page"
	"github.com/microcosm-cc/bluemonday"
)

type pdf struct {
	content  []byte
	settings page.PrintToPDFParams
}

func (p pdf) saveToFile(filename string) {
	err := ioutil.WriteFile(filename, p.content, 0644)
	if err != nil {
		log.Println(err)
	}
	log.Println("PDF saved")
}

func InitServer(port int) {
	log.Println("Starting server on port", port)
	http.HandleFunc("/convert", createPDFHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

func readRequest(r io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(r)
	defer r.Close()
	return body, err
}

func createPDFHandler(w http.ResponseWriter, r *http.Request) {
	if err := validateCreatePDFRequest(w, r); err != nil {
		log.Println(err)
		return
	}

	body, err := readRequest(r.Body)
	body = sanitizeHTMLBody(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runChrome(body)
}

func validateCreatePDFRequest(w http.ResponseWriter, r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")

	if r.Method != "POST" {
		errMsg := "Method not allowed"
		http.Error(w, errMsg, http.StatusMethodNotAllowed)
		return errors.New(errMsg)
	}

	if contentType != "text/plain" && contentType != "text/html" {
		errMsg := "Content-Type must be text/plain or text/html"
		http.Error(w, errMsg, http.StatusBadRequest)
		return errors.New(errMsg)
	}

	if contentLength == "" {
		errMsg := "Content-Length must be set"
		http.Error(w, errMsg, http.StatusBadRequest)
		return errors.New(errMsg)
	}

	return nil
}

func sanitizeHTMLBody(body []byte) []byte {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("html", "head", "title", "body", "style")
	policy.AllowAttrs("style").OnElements("body", "table", "tr", "td", "p", "a", "font", "image")
	policy.AllowAttrs("name").OnElements("meta")
	policy.AllowAttrs("content").OnElements("meta")
	return policy.SanitizeBytes(body)
}
