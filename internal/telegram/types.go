package telegram

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Field struct {
	Name        string
	Type        []string
	Description string
	Required    bool
}

type Type struct {
	Name        string
	Description string
	Fields      []Field
}

func (p *PageAPI) GetType(name string) (Type, error) {
	if len(p.Types) == 0 {
		return Type{}, fmt.Errorf("failed to load types: types map is empty, call LoadTypes() first")
	}

	if typ, exists := p.Types[name]; exists {
		return typ, nil
	}
	return Type{}, fmt.Errorf("type %s not found", name)
}

func (p *PageAPI) GetTypes() (map[string]Type, error) {
	if len(p.Types) == 0 {
		if err := p.LoadTypes(); err != nil {
			return nil, fmt.Errorf("failed to load types: %w", err)
		}
	}
	return p.Types, nil
}

func (p *PageAPI) LoadTypes() error {
	var types []Type
	var currentType Type

	sel := p.Document.Find("h4, table")
	for i := range sel.Nodes {
		s := sel.Eq(i)
		switch {
		case s.Is("h4"):
			if currentType.Name != "" && strings.Contains(currentType.Description, "A placeholder,") {
				types = append(types, currentType)
			}
			if currentType.Name != "" && (len(currentType.Fields) > 0 || isUnionType(currentType.Description)) {
				types = append(types, currentType)
			}
			currentType = Type{Name: strings.TrimSpace(s.Text())}

			nextSibling := s.Next()
			for nextSibling.Length() > 0 && !nextSibling.Is("table") && !nextSibling.Is("h4") {
				if nextSibling.Is("p") {
					currentType.Description += nextSibling.Text() + "\n"
					ul := nextSibling.Next()
					if ul.Is("ul") {
						ul.Find("li a").Each(func(_ int, a *goquery.Selection) {
							if href, exists := a.Attr("href"); exists && strings.HasPrefix(href, "#") {
								currentType.Description += "- " + a.Text() + "\n"
							}
						})
					}
				}
				nextSibling = nextSibling.Next()
			}
		case s.Is("table"):
			if currentType.Name == "" {
				continue
			}

			firstHeader := strings.TrimSpace(s.Find("thead th").First().Text())
			if firstHeader != "Field" {
				break
			}

			s.Find("tbody > tr").Each(func(_ int, tr *goquery.Selection) {
				var field Field
				var isRequired bool
				tr.Find("td").Each(func(j int, td *goquery.Selection) {
					header := tr.Parent().Prev().Find("th").Eq(j).Text()
					switch strings.TrimSpace(header) {
					case "Field":
						field.Name = td.Text()
					case "Type":
						types := strings.Split(td.Text(), " or ")
						field.Type = make([]string, len(types))
						for i, t := range types {
							field.Type[i] = strings.TrimSpace(t)
						}
					case "Required":
						isRequired = td.Text() == "Yes"
					case "Description":
						field.Description = td.Text()
					default:
					}
				})
				field.Required = isRequired
				currentType.Fields = append(currentType.Fields, field)
			})
		}
	}
	if currentType.Name != "" && strings.Contains(currentType.Description, "A placeholder,") {
		types = append(types, currentType)
	}
	if currentType.Name != "" && ((len(currentType.Fields) > 0) || (len(currentType.Fields) == 0 && isUnionType(currentType.Description))) {
		types = append(types, currentType)
	}

	for i := range types {
		p.Types[strings.ToLower(types[i].Name)] = types[i]
	}

	return nil
}

func isUnionType(desc string) bool {
	desc = strings.ToLower(desc)
	return strings.Contains(desc, "this object represents") || strings.Contains(desc, "this object describes") || strings.Contains(desc, "this object contains") || strings.Contains(desc, "can be one of")
}