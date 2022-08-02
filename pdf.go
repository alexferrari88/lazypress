// lazypress converts HTML pages to PDF looking just like they would render in the browser.
// Under-the-hood, it is using Google Chrome (via [github.com/chromedp/chromedp]) to load the HTML so the PDF looks just like if you would print it from the browser.
// It also comes with a built-in HTML sanitizer (uses [github.com/microcosm-cc/bluemonday]).
// You can use this code as a library, or you can run it as a server which will spit out PDFs when you send HTML to it.
package lazypress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/microcosm-cc/bluemonday"
)

// PDF represents a PDF document as it is used by lazypress.
type PDF struct {
	Content  []byte
	Settings page.PrintToPDFParams
	Exporter io.Writer
	Closer   io.Closer
	filePath string
	Sanitize bool
}

// Export outputs the generated PDF to the configured output.
// See LoadSettings to configure the output type.
func (p PDF) Export() error {
	if p.Exporter == nil {
		return fmt.Errorf("no exporter set")
	}
	_, err := p.Exporter.Write(p.Content)
	if err != nil {
		return fmt.Errorf("could not export PDF: %v", err)
	}
	log.Println("PDF exported")
	if p.filePath != "" {
		log.Println("PDF saved to", p.filePath)
	}
	if p.Closer != nil {
		p.Closer.Close()
	}
	return nil
}

func (p *PDF) createFile(filename string) (io.WriteCloser, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		// falback to tmp dir
		dir = os.TempDir()
	}
	if filename == "" {
		filename = "lazypress"
	}
	if !strings.HasSuffix(filename, "*.pdf") {
		filename = filename + "*.pdf"
	}
	file, err := ioutil.TempFile(dir, filename)
	if err != nil {
		log.Fatal(err)
	}
	p.filePath = file.Name()
	return file, nil
}

// LoadSettings loads a map[string]string of settings to configure the PDF.
// The map can contain the following keys:
// - output: the output type. Can be "file", "download".
// - filename: the filename to use when outputting to a file.
// - sanitize: whether to sanitize the HTML.
// Since we are also using the same settings as the [github.com/chromedp/cdproto/page], you can also use the same keys.
// See https://pkg.go.dev/github.com/chromedp/cdproto/page#PrintToPDFParams for more information.
func (p *PDF) LoadSettings(params map[string]string, w io.Writer, c io.Closer) error {
	if err := queryParamsToStruct(params, &p.Settings, "json"); err != nil {
		return err
	}
	if strings.ToLower(params["sanitize"]) == "true" {
		p.Sanitize = true
		if p.Settings.HeaderTemplate != "" {
			p.Settings.HeaderTemplate = string(SanitizeHTML([]byte(p.Settings.HeaderTemplate)))
		}
		if p.Settings.FooterTemplate != "" {
			p.Settings.FooterTemplate = string(SanitizeHTML([]byte(p.Settings.FooterTemplate)))
		}
	}
	outputType := strings.ToLower(params["output"])
	switch outputType {
	case "file":
		file, err := p.createFile(params["filename"])
		if err != nil {
			p.Exporter = w
			p.Closer = c
			return nil
		}
		p.Exporter = file
		p.Closer = file
	case "download":
		p.Exporter = w
		p.Closer = c
	case "s3":
		// TODO: implement
		p.Exporter = w
		p.Closer = c
	case "email":
		// TODO: implement
		p.Exporter = w
		p.Closer = c
	default:
		if w != nil {
			p.Exporter = w
			p.Closer = c
		} else {
			p.Exporter = os.Stdout
			p.Closer = os.Stdout
		}
	}
	return nil
}

// SanitizeHTML sanitizes the HTML using [github.com/microcosm-cc/bluemonday].
func SanitizeHTML(c []byte) []byte {
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
