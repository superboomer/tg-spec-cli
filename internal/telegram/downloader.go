package telegram

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// httpClient is used to fetch documentation pages. It enforces a timeout so a
// slow or unresponsive server can't hang the CLI indefinitely.
var httpClient = &http.Client{Timeout: 30 * time.Second}

type PageAPI struct {
	Types    map[string]Type
	Document *goquery.Document
}

func GetPage(urlStr string) (*PageAPI, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("URL must have a host")
	}

	res, err := httpClient.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return &PageAPI{Document: doc, Types: make(map[string]Type)}, nil
}
