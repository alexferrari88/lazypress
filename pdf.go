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
	content  []byte
	Settings page.PrintToPDFParams
	exporter io.Writer
	closer   io.Closer
}

func (p PDF) Export() error {
	if p.exporter == nil {
		return fmt.Errorf("No exporter set")
	}
	_, err := p.exporter.Write(p.content)
	if err != nil {
		return fmt.Errorf("Could not export PDF: %v", err)
	}
	log.Println("PDF exported")
	if p.closer != nil {
		p.closer.Close()
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
		filename = "lazypress*.pdf"
	}
	file, err := ioutil.TempFile(dir, filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created file", file.Name())
	return file, nil
}

func (p *PDF) loadSettings(params map[string]string, w io.Writer) error {
	if err := queryParamsToStruct(params, &p.Settings, "json"); err != nil {
		return err
	}
	outputType := strings.ToLower(params["outputType"])
	switch outputType {
	case "file":
		file, err := p.createFile(params["filename"])
		if err != nil {
			p.exporter = w
			return nil
		}
		p.exporter = file
		p.closer = file
	case "download":
		p.exporter = w
	case "s3":
		// TODO: implement
		p.exporter = w
	case "email":
		// TODO: implement
		p.exporter = w
	default:
		p.exporter = w
	}
	return nil
}
