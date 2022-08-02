package lazypress

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// InitServer initializes the server.
// It takes the port to listen on and the path to the chrome executable.
// If the chrome executable is not provided, the server will use the default options of [github.com/chromedp/chromedp].
// The default port is 3444.
func InitServer(port int, chromePath string) {
	log.Println("Starting server on port", port)
	http.HandleFunc("/convert", convertHTMLServerHandler(chromePath))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

func readRequest(r io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(r)
	defer r.Close()
	return body, err
}

func convertHTMLServerHandler(chromePath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := validateConvertHTMLRequest(w, r); err != nil {
			log.Println(err)
			return
		}
		var p PDF

		params := urlQueryToMap(r.URL.Query())
		if err := p.LoadSettings(params, w, nil); err != nil {
			// we just log the error and continue with defaults
			log.Println(err)
			p.Settings = page.PrintToPDFParams{}
		}
		body, err := readRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			http.Error(w, "Body is empty", http.StatusBadRequest)
			return
		}
		if p.Sanitize {
			body = SanitizeHTML(body)
			if len(body) == 0 {
				http.Error(w, "Body is empty", http.StatusBadRequest)
				return
			}
		}

		var allocatorCtx context.Context
		var allocatorCancel context.CancelFunc

		if chromePath != "" {
			opt := []func(allocator *chromedp.ExecAllocator){
				chromedp.ExecPath(chromePath),
			}
			// create context
			allocatorCtx, allocatorCancel = chromedp.NewExecAllocator(
				context.Background(),
				append(opt, chromedp.DefaultExecAllocatorOptions[:]...)[:]...,
			)
			defer allocatorCancel()
		} else {
			allocatorCtx = context.Background()
		}

		p.GenerateWithChrome(allocatorCtx, body)
		if p.Content == nil {
			log.Println("Could not generate PDF")
			http.Error(w, "Could not generate PDF", http.StatusInternalServerError)
			return
		}
		if err := p.Export(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p.filePath != "" {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf("{\"file\": \"%s\"}", p.filePath)))
		}
	}
}

func urlQueryToMap(query url.Values) map[string]string {
	params := make(map[string]string, len(query))
	for k, v := range query {
		params[k] = v[0]
	}
	return params
}

func validateConvertHTMLRequest(w http.ResponseWriter, r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")

	if r.Method != "POST" {
		errMsg := "method not allowed"
		http.Error(w, errMsg, http.StatusMethodNotAllowed)
		return errors.New(errMsg)
	}

	if contentType != "text/plain" && contentType != "text/html" {
		errMsg := "content-type must be text/plain or text/html"
		http.Error(w, errMsg, http.StatusBadRequest)
		return errors.New(errMsg)
	}

	if contentLength == "" || contentLength == "0" {
		errMsg := "content-length must be set"
		http.Error(w, errMsg, http.StatusBadRequest)
		return errors.New(errMsg)
	}

	return nil
}
