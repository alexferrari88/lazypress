package lazypress

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/chromedp/cdproto/page"
)

type PDF struct {
	HTMLContent []byte
	Settings    page.PrintToPDFParams
	Exporter    io.Writer
	Closer      io.Closer
	filePath    string
	Sanitize    bool
}

func (p PDF) Export() error {
	if p.Exporter == nil {
		return fmt.Errorf("no exporter set")
	}
	_, err := p.Exporter.Write(p.HTMLContent)
	if err != nil {
		return fmt.Errorf("could not export PDF: %v", err)
	}
	log.Println("PDF exported")
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
	log.Println("Created file", p.filePath)
	return file, nil
}

func (p *PDF) loadSettings(params map[string]string, w io.Writer, c io.Closer) error {
	if err := queryParamsToStruct(params, &p.Settings, "json"); err != nil {
		return err
	}
	if strings.ToLower(params["sanitize"]) == "true" {
		p.Sanitize = true
		if p.Settings.HeaderTemplate != "" {
			p.Settings.HeaderTemplate = string(sanitizeHTML([]byte(p.Settings.HeaderTemplate)))
		}
		if p.Settings.FooterTemplate != "" {
			p.Settings.FooterTemplate = string(sanitizeHTML([]byte(p.Settings.FooterTemplate)))
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
