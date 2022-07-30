package lazypress

import (
	"context"
	"io"
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
)

func (p *PDF) GenerateWithChrome(html []byte) *PDF {
	var wg sync.WaitGroup

	// locate chrome executable path
	dir, dirError := os.Getwd()
	if dirError != nil {
		panic(dirError)
	}
	opt := []func(allocator *chromedp.ExecAllocator){
		chromedp.ExecPath(path.Join(dir, "../chrome-linux", "chrome")),
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
					buf, _, err := p.Settings.Do(ctx)
					if err != nil {
						return err
					}
					p.content = buf
					log.Println("PDF content created")
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

	return p

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
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := emulation.SetDeviceMetricsOverride(1920, 1080, 0, false).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
	}
}
