﻿package lazypress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
)

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
			body = sanitizeHTML(body)
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

func sanitizeHTML(c []byte) []byte {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("html", "head", "title", "body", "style")
	policy.AllowAttrs("style").OnElements("body", "table", "tr", "td", "p", "a", "font", "image")
	policy.AllowAttrs("name").OnElements("meta")
	policy.AllowAttrs("content").OnElements("meta")
	return policy.SanitizeBytes(c)
}

func queryParamsToStruct(params map[string]string, structToUse any, tagStr string) error {
	// From https://medium.com/wesionary-team/reflections-tutorial-query-string-to-struct-parser-in-go-b2f858f99ea1

	var err error
	dType := reflect.TypeOf(structToUse)
	if dType.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}
	dValue := reflect.ValueOf(structToUse)
	for i := 0; i < dType.Elem().NumField(); i++ {
		field := dType.Elem().Field(i)
		key := strings.Replace(field.Tag.Get(tagStr), ",omitempty", "", -1)
		kind := field.Type.Kind()

		settingVal, ok := params[key]
		if !ok {
			continue
		}

		fieldVal := dValue.Elem().Field(i)

		switch kind {
		case reflect.String:
			if fieldVal.CanSet() {
				fieldVal.SetString(settingVal)
			}
		case reflect.Int:
			intVal, err := strconv.ParseInt(settingVal, 10, 64)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(intVal)
			}
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(settingVal)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetBool(boolVal)
			}
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(settingVal, 64)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetFloat(floatVal)
			}
		case reflect.Struct:
			if fieldVal.CanSet() {
				val := reflect.New(field.Type)
				err := json.Unmarshal([]byte(settingVal), val.Interface())
				if err != nil {
					return err
				}
				fieldVal.Set(val.Elem())
			}
		}
	}
	return err
}
