![Logo](https://github.com/alexferrari88/lazypress/blob/master/logo.png?raw=true)

# lazypress

Convert HTML pages to PDF looking just like they would render in the browser.

Under-the-hood, lazypress is using Google Chrome (via [chromedp](https://github.com/chromedp/chromedp)) to load the HTML so the PDF looks just like if you would print it from the browser.

It also comes with a built-in HTML sanitizer (uses [bluemonday](https://github.com/microcosm-cc/bluemonday)).

You can use this code as a library, or you can run it as a server which will spit out PDFs when you send HTML to it.

## Features 🚀

- Sanitize HTML to remove potentially malicious code
- Tweak parameters (landscape/portrait, page size, content scale, margins, etc. )to get a PDF like you want it
- Add custom header and footer
- Save to local file or return it as PDF to be downloaded (`Content-type: application/pdf`)
- Run with your own Chrome or within a Docker container
- Use as server if you want just a turnkey solution or as a Go library to include it in your applications

## How to use ⚙️

### As a server

Start the server with:

```bash
lazypress
```

By default, the server is listening on port `3444`. You can change the port by running:

```bash
lazypress --port PORT
```

You can also pass a custom path for Google chrome:

```bash
lazypress --chrome CHROME_PATH
```

Once the server is started, you can send POST requests to the `/convert` endpoint.

#### Request

- Method: POST
- Content-type: text/plain or text/html

You can tweak the settings of the PDF and decide the output location by passing specific query parameters.

_Unfortunately, at least until I find a solution (or you do and send a PR 😉), these query parameters are case sensitive._

Query parameters:

- `sanitize`
  - If true, the server will clean up the HTML to remove potentially malicious code.
  - options: true | none
  - default: none
- `output`
  - Specify where to output the generated PDF
  - options: file | download | none
  - default: download
- `filename`
  - If output is set to "file", this allows you to choose a file name for the PDF _(I might remove this due to security concerns)_
- `landscape`
  - Paper orientation
  - options: true | false
  - default: false
- `displayHeaderFooter`
  - Display header and footer
  - options: true | false
  - default: false
- `printBackground`
  - Print background graphics
  - options: true | false
  - default: false
- `scale`
  - Scale of the webpage rendering (float)
  - default: 1
- `paperWidth`
  - Paper width in inches (float)
  - default: 8.5
- `paperHeight`
  - Paper height in inches (float)
  - default: 11
- `marginTop`
  - Top margin in inches (float)
  - default: 0.4
- `marginBottom`
  - Bottom margin in inches (float)
  - default: 0.4
- `marginLeft`
  - Left margin in inches (float)
  - default: 0.4
- `marginRight`
  - Right margin in inches (float)
  - default: 0.4
- `pageRanges`
  - Paper ranges to print, one based, e.g., '1-5, 8, 11-13'. Pages are printed in the document order, not in the order specified, and no more than once. Defaults to empty string, which implies the entire document is printed. The page numbers are quietly capped to actual page count of the document, and ranges beyond the end of the document are ignored. If this results in no pages to print, an error is reported. It is an error to specify a range with start greater than end.
- `headerTemplate`
  - HTML template for the print header. Should be valid HTML markup with following classes used to inject printing values into them: - date: formatted print date - title: document title - url: document location - pageNumber: current page number - totalPages: total pages in the document For example, `<span class=title></span>` would generate span containing the title.
- `footerTemplate`
  - HTML template for the print footer. Should use the same format as the headerTemplate.
- `preferCSSPageSize`
  - Whether or not to prefer page size as defined by css. Defaults to false, in which case the content will be scaled to fit the paper size.

### As a library

Refer to the [GoDoc](https://pkg.go.dev/github.com/alexferrari88/lazypress).

## FAQ 🤔

#### Can I use it as serverless (e.g. with AWS Lambda)?

I'm working on it. Check the branch feature/serverless.

Contributions are more than welcome!

#### Can I have the PDF sent via email instead?

I'm working on it. If you want to implement it yourself, please add the implementation in the `LoadSettings` function.

#### Can I have the PDF loaded on S3 instead?

Please, refer to the previous question.

## Acknowledgements 🙏🏼

- [https://github.com/chromedp/chromedp](https://github.com/chromedp/chromedp): for providing headless Chrome
- [https://github.com/microcosm-cc/bluemonday/](https://github.com/microcosm-cc/bluemonday/): for providing an HTML sanitizer
- my wife: for bearing with me 🤗

## Contributing 🤝🏼

Contributions are always welcome! Just send a PR and after a review, I will be glad to merge your changes!
