package telegram

import (
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (p *PageAPI) GetVersion() (string, error) {
	var version string
	var foundRecentChanges bool

	p.Document.Find("strong").Each(func(_ int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Bot API") {
			if version != "" {
				return
			}
			version = strings.TrimPrefix(s.Text(), "Bot API ")
		}
	})

	if version != "" {
		return version, nil
	}

	p.Document.Find("*").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if !foundRecentChanges && s.Is("h3") && strings.Contains(strings.ToLower(s.Text()), "recent changes") {
			foundRecentChanges = true
			return true // continue
		}
		if foundRecentChanges && s.Is("h4") {
			version = strings.TrimSpace(s.Text())
			return false // stop
		}
		return true // continue
	})

	if !foundRecentChanges {
		return "", errors.New("can't find 'Recent Changes' section")
	}
	if version == "" {
		return "", errors.New("can't find <h4> tag after 'Recent Changes'")
	}
	return version, nil
}
