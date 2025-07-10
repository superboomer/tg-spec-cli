package telegram

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PageAPI struct {
	Types    map[string]Type
	Document *goquery.Document
}

func GetPage(url string) (*PageAPI, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	return &PageAPI{Document: doc, Types: make(map[string]Type)}, nil
}
