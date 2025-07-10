package telegram

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DataType struct {
	Types      []string
	IsArray    bool
	ArrayDepth int
}

func (p *PageAPI) parseDataType(doc *goquery.Selection) DataType {
	var dataType DataType

	fullText := doc.Text()
	for strings.Contains(fullText, "Array of") {
		dataType.IsArray = true
		dataType.ArrayDepth++
		fullText = strings.Replace(fullText, "Array of", "", 1)
	}

	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && strings.HasPrefix(href, "#") {
			typeName := strings.TrimPrefix(href, "#")

			typeData, err := p.GetType(typeName)
			if err == nil {
				if typeName != "" {
					dataType.Types = append(dataType.Types, typeData.Name)
				}
			}
		}
	})

	if len(dataType.Types) == 0 {
		text := doc.Text()
		for strings.Contains(text, "Array of") {
			text = strings.Replace(text, "Array of", "", 1)
		}
		types := strings.Split(text, " or ")
		for _, t := range types {
			t = strings.TrimSpace(t)
			if t != "" {
				dataType.Types = append(dataType.Types, t)
			}
		}
	}

	return dataType
}
