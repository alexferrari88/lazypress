package lazypress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/microcosm-cc/bluemonday"
)

type PDF struct {
	content  []byte
	Settings page.PrintToPDFParams
}

func (p PDF) saveToFile(filename string) {
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
	var p PDF

	if err := loadPDFSettings(r, &p.Settings); err != nil {
		// we just log the error and continue with defaults
		log.Println(err)
		p.Settings = page.PrintToPDFParams{}
	}
	body, err := readRequest(r.Body)
	body = sanitizeHTMLBody(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.FromChrome(body).saveToFile("test.pdf")
}

func loadPDFSettings(r *http.Request, settings *page.PrintToPDFParams) error {
	if err := queryParamsToStruct(r, settings, "json"); err != nil {
		return err
	}
	return nil
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

func queryParamsToStruct(r *http.Request, d any, tagStr string) error {
	// From https://medium.com/wesionary-team/reflections-tutorial-query-string-to-struct-parser-in-go-b2f858f99ea1
	var err error
	dType := reflect.TypeOf(d)
	if dType.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}
	dValue := reflect.ValueOf(d)
	for i := 0; i < dType.Elem().NumField(); i++ {
		field := dType.Elem().Field(i)
		key := strings.Replace(field.Tag.Get(tagStr), ",omitempty", "", -1)
		kind := field.Type.Kind()

		queryParam := r.URL.Query().Get(key)
		if queryParam == "" {
			continue
		}

		fieldVal := dValue.Elem().Field(i)

		switch kind {
		case reflect.String:
			if fieldVal.CanSet() {
				fieldVal.SetString(queryParam)
			}
		case reflect.Int:
			intVal, err := strconv.ParseInt(queryParam, 10, 64)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(intVal)
			}
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(queryParam)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetBool(boolVal)
			}
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(queryParam, 64)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetFloat(floatVal)
			}
		case reflect.Struct:
			if fieldVal.CanSet() {
				val := reflect.New(field.Type)
				err := json.Unmarshal([]byte(queryParam), val.Interface())
				if err != nil {
					return err
				}
				fieldVal.Set(val.Elem())
			}
		}
	}
	return err
}
