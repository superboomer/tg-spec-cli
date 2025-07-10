package telegram

import (
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type ReturnType struct {
	Name    string
	IsArray bool
}

type Parameter struct {
	Name        string
	Type        DataType
	Description string
	Required    bool
}

type Method struct {
	ReturnType  ReturnType
	Name        string
	Description string
	Parameters  []Parameter
}

func (p *PageAPI) GetMethods() ([]Method, error) {
	var methods []Method
	var currentMethod Method

	sel := p.Document.Find("h4, table")
	for i := range sel.Nodes {
		s := sel.Eq(i)
		switch {
		case s.Is("h4"):
			if currentMethod.Name != "" && len(currentMethod.Parameters) > 0 {
				methods = append(methods, currentMethod)
			}
			currentMethod = Method{Name: strings.TrimSpace(s.Text())}

			nextSibling := s.Next()
			for nextSibling.Length() > 0 && !nextSibling.Is("table") && !nextSibling.Is("h4") {
				if nextSibling.Is("p") {
					currentMethod.Description += nextSibling.Text()

					var fullText string
					var returnTypeName string
					var isArray bool

					nextSibling.Contents().Each(func(_ int, s *goquery.Selection) {
						fullText += s.Text()

						if strings.Contains(s.Text(), "array of") {
							isArray = true
						}
						if goquery.NodeName(s) == "em" {
							returnTypeName = s.Text()
							fullText += returnTypeName
						} else if goquery.NodeName(s) == "a" {

							if href, exists := s.Attr("href"); exists {

								if isFirstLetterUppercase(s.Text()) {
									i := strings.Index(href, "#")
									if i != -1 {
										href = href[i+1:]
									}

									typeData, err := p.GetType(href)
									if err == nil {
										returnTypeName = typeData.Name
										fullText += returnTypeName
									}
								}
							}
						}
					})

					if currentMethod.ReturnType.Name == "" || !isFirstLetterUppercase(currentMethod.ReturnType.Name) {
						switch {
						case returnTypeName == "True":
							currentMethod.ReturnType = ReturnType{Name: "boolean", IsArray: isArray}
						case returnTypeName == "Int":
							currentMethod.ReturnType = ReturnType{Name: "integer", IsArray: isArray}
						default:
							typeData, err := p.GetType(strings.ToLower(returnTypeName))
							if err == nil {
								currentMethod.ReturnType = ReturnType{Name: typeData.Name, IsArray: isArray}
							}
						}
					}
				}
				nextSibling = nextSibling.Next()
			}

		case s.Is("table"):
			if currentMethod.Name == "" {
				continue
			}

			firstHeader := strings.TrimSpace(s.Find("thead th").First().Text())
			if firstHeader != "Parameter" {
				break
			}

			s.Find("tbody > tr").Each(func(_ int, tr *goquery.Selection) {
				var parameter Parameter
				var isRequired bool
				tr.Find("td").Each(func(j int, td *goquery.Selection) {
					header := tr.Parent().Prev().Find("th").Eq(j).Text()
					switch strings.TrimSpace(header) {
					case "Parameter":
						parameter.Name = td.Text()
					case "Type":
						parameter.Type = p.parseDataType(td)
					case "Required":
						isRequired = td.Text() == "Yes"
					case "Description":
						parameter.Description = td.Text()
					default:
					}
				})
				parameter.Required = isRequired
				currentMethod.Parameters = append(currentMethod.Parameters, parameter)
			})
		}
	}

	if currentMethod.Name != "" && len(currentMethod.Parameters) > 0 {
		methods = append(methods, currentMethod)
	}

	return methods, nil
}

func isFirstLetterUppercase(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	first := runes[0]
	return unicode.IsLetter(first) && unicode.IsUpper(first)
}
