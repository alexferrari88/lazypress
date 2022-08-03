package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log"

	"github.com/alexferrari88/lazypress"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(r Request) (Response, error) {
	if statusCode, err := validateConvertHTMLRequest(r); err != nil {
		return makeResponse(statusCode, err.Error(), false, nil), nil
	}

	var p lazypress.PDF
	params := r.QueryStringParameters
	if err := p.LoadSettings(params, nil, nil); err != nil {
		// we just log the error and continue with defaults
		log.Println(err)
		p.Settings = page.PrintToPDFParams{}
	}
	body := []byte(r.Body)
	errorMsg := "Request body is empty"

	if len(body) == 0 {
		return makeResponse(400, errorMsg, false, nil), nil
	}
	if p.Sanitize {
		body = sanitizeHTML(body)
		if len(body) == 0 {
			return makeResponse(400, errorMsg, false, nil), nil
		}
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Headless,
		chromedp.NoSandbox,
		// chromedp.WindowSize(1200, 1000),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
	}
	// create context
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(opts, chromedp.DefaultExecAllocatorOptions[:]...)[:]...,
	)
	defer allocatorCancel()
	p.GenerateWithChrome(allocatorCtx, body)
	if p.Content == nil {
		log.Println("Could not generate PDF")
		return makeResponse(500, "Could not generate PDF", false, nil), nil
	}
	if err := p.Export(); err != nil {
		log.Println(err)
		return makeResponse(500, "Could not generate PDF", false, nil), nil
	}
	headers := map[string]string{
		"Content-Type":        "application/pdf",
		"Content-Disposition": "attachment; filename=lazypress.pdf",
	}
	return makeResponse(200, "All good", base64.StdEncoding.EncodeToString(p.Content), headers), nil
}

func main() {
	lambda.Start(Handler)
}

func makeResponse(statusCode int, body string, isBase64Encoded bool, headers map[string]string) Response {
	if headers == nil {
		headers = map[string]string{
			"Content-Type": "application/json",
		}
	}
	return Response{
		StatusCode:      statusCode,
		IsBase64Encoded: isBase64Encoded,
		Body:            body,
		Headers:         headers,
	}
}

func validateConvertHTMLRequest(r Request) (int, error) {
	if contentLength, found := r.Headers["Content-Length"]; !found {
		return 400, errors.New("content-length must be set")
	} else if contentLength == "" || contentLength == "0" {
		return 400, errors.New("content-length must be set")
	}

	contentType := r.Headers["Content-Type"]
	if r.HTTPMethod != "POST" {
		return 405, errors.New("method not allowed")
	}

	if contentType != "text/plain" && contentType != "text/html" {
		return 400, errors.New("content-type must be text/plain or text/html")
	}

	return 0, nil
}

func sanitizeHTML(c []byte) []byte {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("html", "head", "title", "body", "style")
	policy.AllowAttrs("style").OnElements("body", "table", "tr", "td", "p", "a", "font", "image")
	policy.AllowAttrs("name").OnElements("meta")
	policy.AllowAttrs("content").OnElements("meta")
	return policy.SanitizeBytes(c)
}
