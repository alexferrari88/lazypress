package lazypress

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func (p *PDF) GenerateWithChrome(html []byte, execPath string) *PDF {
	var wg sync.WaitGroup
	var allocatorCtx context.Context
	var allocatorCancel context.CancelFunc

	if execPath != "" {
		opt := []func(allocator *chromedp.ExecAllocator){
			chromedp.ExecPath(execPath),
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
					p.HTMLContent = buf
					log.Println("PDF content created")
					return nil
				})); err != nil {
					log.Println(err)
				}

			}()

		}
	})

	// save the HTML content to a temporary file
	htmlFile, err := ioutil.TempFile("", "lazypress*.html")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(htmlFile.Name())
	defer htmlFile.Close()
	if _, err := htmlFile.Write(html); err != nil {
		log.Fatalln(err)
	}

	// start browser and load html
	if err := chromedp.Run(ctx, loadHTMLInBrowser(html, htmlFile.Name())); err != nil {
		log.Fatalln(err)
	}

	wg.Wait()

	return p

}

func loadHTMLInBrowser(html []byte, fileName string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := emulation.SetScriptExecutionDisabled(true).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
		chromedp.Navigate(fmt.Sprintf("file://%s", fileName)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := emulation.SetDeviceMetricsOverride(1920, 1080, 0, false).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
	}
}
