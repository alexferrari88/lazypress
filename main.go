package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
)

func main() {
	InitServer(3444)
}

type settings struct {
	sanitize bool
}

type pdf struct {
	content []byte
	settings
}

func (p pdf) saveToFile(filename string) {
	err := ioutil.WriteFile(filename, p.content, 0644)
	if err != nil {
		log.Println(err)
	}
	log.Println("PDF saved")
}

func runChrome(html []byte) {
	var wg sync.WaitGroup
	var p pdf

	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		panic(dirError)
	}
	opt := []func(allocator *chromedp.ExecAllocator){
		chromedp.ExecPath(path.Join(dir, "chrome-linux", "chrome")),
	}

	// create context
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(opt, chromedp.DefaultExecAllocatorOptions[:]...)[:]...,
	)
	defer allocatorCancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	// add a listener for when the page is fully loaded
	// this allows us to give the page time to render the images as well
	wg.Add(1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev.(type) {
		case *page.EventLoadEventFired:
			go func() {
				defer wg.Done()

				// create the pdf
				if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
					if err != nil {
						return err
					}
					p.content = buf
					log.Println("PDF created")
					return nil
				})); err != nil {
					log.Println(err)
				}

			}()

		}
	})

	// create test server with the html we passed in
	ts := httptest.NewServer(writeHTML(html))
	defer ts.Close()

	// start browser and load html
	if err := chromedp.Run(ctx, loadHTMLInBrowser(html, ts)); err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	p.saveToFile("test.pdf")

}

func writeHTML(content []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(string(content)))
	})
}

func loadHTMLInBrowser(html []byte, ts *httptest.Server) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := emulation.SetScriptExecutionDisabled(true).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
		chromedp.Navigate(ts.URL),
	}
}

func InitServer(port int) {
	log.Println("Starting server on port", port)
	http.HandleFunc("/convert", createPDFHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

func createPDFHandler(w http.ResponseWriter, r *http.Request) {
	if err := validateCreatePDFRequest(w, r); err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	body = sanitizeBody(body)
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

func sanitizeBody(body []byte) []byte {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("html", "head", "title", "body", "style")
	policy.AllowAttrs("style").OnElements("body", "table", "tr", "td", "p", "a", "font", "image")
	policy.AllowAttrs("name").OnElements("meta")
	policy.AllowAttrs("content").OnElements("meta")
	return policy.SanitizeBytes(body)
}
