package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

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
		log.Fatal(err)
	}
}

func runChrome(html []byte) {
	dir, dirError := os.Getwd()
	if dirError != nil {
		panic(dirError)
	}
	opt := []func(allocator *chromedp.ExecAllocator){
		chromedp.ExecPath(path.Join(dir, "chrome-linux", "chrome")),
	}

	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(opt, chromedp.DefaultExecAllocatorOptions[:]...)[:]...,
	)
	defer allocatorCancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	var p pdf
	if err := chromedp.Run(ctx, printToPDF(html, &p.content)); err != nil {
		log.Fatal(err)
	}

	p.saveToFile("test.pdf")

}

func printToPDF(html []byte, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := emulation.SetScriptExecutionDisabled(true).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, string(html)).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func InitServer(port int) {
	fmt.Println("Starting server on port", port)
	http.HandleFunc("/convert", createPDFHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server started on port", port)
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
	fmt.Println(string(body))
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
	policy.AllowElements("html", "head", "meta", "title", "body", "style")
	policy.AllowAttrs("style").OnElements("body", "table", "tr", "td", "p", "a", "font", "image")
	return policy.SanitizeBytes(body)
}
