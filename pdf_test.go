package lazypress

import (
	"os"
	"strings"
	"testing"
)

type mockWriter struct {
	written []byte
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.written = p
	return len(p), nil
}

type mockCloser struct {
	count uint8
}

func (m *mockCloser) Close() error {
	m.count++
	return nil
}

func TestShouldErrorWhenExporterIsNotDefined(t *testing.T) {
	var p PDF
	err := p.Export()
	if err == nil {
		t.Error("Expected error when exporter is not defined")
	}
}

func TestShouldWriteWithExporter(t *testing.T) {
	var p PDF
	m := &mockWriter{}
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

func TestShouldCreatePDFWithLazypressInFilename(t *testing.T) {
	var p PDF
	f, err := p.createFile("")
	if err != nil {
		t.Error("Expected no error when creating file")
	}
	filePath := f.(*os.File).Name()
	defer os.Remove(filePath)
	if !strings.Contains(filePath, "lazypress") {
		t.Error("Expected filePath to contain lazypress")
	}
	if !strings.HasSuffix(filePath, ".pdf") {
		t.Error("Expected filePath to have .pdf suffix")
	}
	if p.filePath != filePath {
		t.Error("Expected filePath to be the same as the created file")
	}
}

func TestShouldCreatePDFNotWithLazypressInFilename(t *testing.T) {
	var p PDF
	fileName := "test"
	f, err := p.createFile(fileName)
	if err != nil {
		t.Error("Expected no error when creating file")
	}
	filePath := f.(*os.File).Name()
	defer os.Remove(filePath)
	if !strings.Contains(filePath, "test") {
		t.Error("Expected filePath to contain test")
	}
	if !strings.HasSuffix(filePath, ".pdf") {
		t.Error("Expected filePath to have .pdf suffix")
	}
	if p.filePath != filePath {
		t.Error("Expected filePath to be the same as the created file")
	}
}

func TestShouldLoadOutputToFileSetting(t *testing.T) {
	var p PDF
	params := map[string]string{
		"output": "file",
	}
	p.loadSettings(params, nil, nil)
	if p.Exporter.(*os.File) == nil {
		t.Error("Expected Exporter to be a file")
	}
	if p.Closer.(*os.File) == nil {
		t.Error("Expected Closer to be a file")
	}
	if p.filePath == "" {
		t.Error("Expected filePath to be set")
	}
	defer os.Remove(p.filePath)
}

func TestShouldLoadOutputToDownloadSetting(t *testing.T) {
	var p PDF
	w := &mockWriter{}
	params := map[string]string{
		"output": "download",
	}
	p.loadSettings(params, w, nil)
	if p.Exporter.(*mockWriter) == nil {
		t.Error("Expected Exporter to be a mockWriter")
	}
	if p.Exporter != w {
		t.Error("Expected Exporter to be the same as the mockWriter")
	}
	if p.filePath != "" {
		t.Error("Expected filePath to be empty")
	}
}

func TestShouldLoadPassedWriterAsDefaultExporter(t *testing.T) {
	var p PDF
	w := &mockWriter{}
	c := &mockCloser{}
	p.loadSettings(map[string]string{}, w, c)
	if p.Exporter.(*mockWriter) != w {
		t.Error("Expected Exporter to be the same as the mockWriter")
	}
	if p.Closer.(*mockCloser) != c {
		t.Error("Expected Closer to be the same as the mockCloser")
	}
	if p.filePath != "" {
		t.Error("Expected filePath to be empty")
	}
}

func TestShouldSetStdoutAsWriterWhenNoWriterPassedAndNoOutputParamPassed(t *testing.T) {
	var p PDF
	p.loadSettings(map[string]string{}, nil, nil)
	if p.Exporter.(*os.File) == nil {
		t.Error("Expected Exporter to be a file")
	}
	if p.Closer.(*os.File) == nil {
		t.Error("Expected Closer to be a file")
	}
	if p.filePath != "" {
		t.Error("Expected filePath to be empty")
	}
}

func TestShouldLoadPassedSettings(t *testing.T) {
	var p PDF
	params := map[string]string{
		"landscape":           "true",
		"displayHeaderFooter": "true",
		"printBackground":     "true",
		"scale":               "2",
		"paperWidth":          "8.5",
		"paperHeight":         "11",
		"marginTop":           "1",
		"marginBottom":        "1",
		"marginLeft":          "1",
		"marginRight":         "1",
		"pageRanges":          "1-5, 8, 11-13",
		"headerTemplate":      "<span class=title></span>",
		"footerTemplate":      "<span class=date></span>",
		"preferCSSPageSize":   "true",
	}
	p.loadSettings(params, nil, nil)
	if p.Settings.Landscape != true {
		t.Error("Expected landscape to be true. Got: ", p.Settings.Landscape)
	}
	if p.Settings.DisplayHeaderFooter != true {
		t.Error("Expected displayHeaderFooter to be true. Got: ", p.Settings.DisplayHeaderFooter)
	}
	if p.Settings.PrintBackground != true {
		t.Error("Expected printBackground to be true. Got: ", p.Settings.PrintBackground)
	}
	if p.Settings.Scale != 2 {
		t.Error("Expected scale to be 2. Got: ", p.Settings.Scale)
	}
	if p.Settings.PaperWidth != 8.5 {
		t.Error("Expected paperWidth to be 8.5. Got: ", p.Settings.PaperWidth)
	}
	if p.Settings.PaperHeight != 11 {
		t.Error("Expected paperHeight to be 11. Got: ", p.Settings.PaperHeight)
	}
	if p.Settings.MarginTop != 1 {
		t.Error("Expected marginTop to be 1. Got: ", p.Settings.MarginTop)
	}
	if p.Settings.MarginBottom != 1 {
		t.Error("Expected marginBottom to be 1. Got: ", p.Settings.MarginBottom)
	}
	if p.Settings.MarginLeft != 1 {
		t.Error("Expected marginLeft to be 1. Got: ", p.Settings.MarginLeft)
	}
	if p.Settings.MarginRight != 1 {
		t.Error("Expected marginRight to be 1. Got: ", p.Settings.MarginRight)
	}
	if p.Settings.PageRanges != "1-5, 8, 11-13" {
		t.Error("Expected pageRanges to be 1-5, 8, 11-13. Got: ", p.Settings.PageRanges)
	}
	if p.Settings.HeaderTemplate != "<span class=title></span>" {
		t.Error("Expected headerTemplate to be <span class=title></span>. Got: ", p.Settings.HeaderTemplate)
	}
	if p.Settings.FooterTemplate != "<span class=date></span>" {
		t.Error("Expected footerTemplate to be <span class=date></span>. Got: ", p.Settings.FooterTemplate)
	}
	if p.Settings.PreferCSSPageSize != true {
		t.Error("Expected preferCSSPageSize to be true. Got: ", p.Settings.PreferCSSPageSize)
	}

}
