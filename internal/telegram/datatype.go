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

	// The <a> scan only captures linked types, but some declarations mix a
	// linked type with a bare primitive (e.g. "InputFile or String"). Recover
	// any primitive alternatives from the text so they aren't dropped.
	recovered := dataType.Types
	for _, part := range strings.Split(stripArrayOf(doc.Text()), " or ") {
		part = strings.TrimSpace(part)
		if isPrimitiveType(part) && !containsString(recovered, part) {
			recovered = append(recovered, part)
		}
	}
	dataType.Types = recovered

	if len(dataType.Types) == 0 {
		text := stripArrayOf(doc.Text())
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

// primitiveTypes are the scalar type names used in the Telegram documentation.
// They are never hyperlinked, so they must be recovered from plain text.
var primitiveTypes = map[string]bool{
	"Integer": true, "Int": true,
	"Float": true, "Double": true,
	"Boolean": true, "Bool": true, "True": true, "False": true,
	"String": true,
}

func isPrimitiveType(s string) bool {
	return primitiveTypes[s]
}

func stripArrayOf(text string) string {
	for strings.Contains(text, "Array of") {
		text = strings.Replace(text, "Array of", "", 1)
	}
	return text
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
