package lazypress_test

import (
	"testing"

	"github.com/alexferrari88/lazypress"
)

type mockExporter struct {
	written []byte
}

func (m *mockExporter) Write(p []byte) (n int, err error) {
	m.written = p
	return len(p), nil
}

func TestShouldErrorWhenExporterIsNotDefined(t *testing.T) {
	var p lazypress.PDF
	err := p.Export()
	if err == nil {
		t.Error("Expected error when exporter is not defined")
	}
}

func TestShouldWriteWithExporter(t *testing.T) {
	var p lazypress.PDF
	m := &mockExporter{}
	p.Exporter = m
	content := []byte("<html><body>Hello World</body></html>")
	p.HTMLContent = content
	if err := p.Export(); err != nil {
		t.Error("Expected no error when exporting")
	}
	if string(m.written) != string(content) {
		t.Error("Expected written content to be the same as the HTML content")
	}
}
