package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/superboomer/tg-spec-cli/internal/openapi"
	"github.com/superboomer/tg-spec-cli/internal/telegram"

	"go.uber.org/zap"
)

type Generator struct {
	log      *zap.Logger
	version  string
	types    map[string]telegram.Type
	methods  []telegram.Method
	typeFlag string
}

func New(log *zap.Logger, version string, types map[string]telegram.Type, methods []telegram.Method) *Generator {
	return &Generator{
		log:     log,
		version: version,
		types:   types,
		methods: methods,
	}
}

func NewWithType(log *zap.Logger, version string, types map[string]telegram.Type, methods []telegram.Method, typeFlag string) *Generator {
	return &Generator{
		log:      log,
		version:  version,
		types:    types,
		methods:  methods,
		typeFlag: typeFlag,
	}
}

func (g *Generator) Generate() (*openapi.OpenAPI, error) {
	g.log.Debug("starting OpenAPI generation", zap.String("version", g.version), zap.String("type", g.typeFlag))

	var info openapi.Info
	var servers []openapi.Server

	if g.typeFlag == "gateway" {
		info = openapi.Info{
			Title:       "Telegram Gateway API",
			Description: `The Gateway API is an HTTP-based interface for phone number verification and related operations. See https://core.telegram.org/gateway/api for details.`,
			Version:     g.version,
		}
		servers = []openapi.Server{
			{
				URL:         "https://gatewayapi.telegram.org/",
				Description: "Telegram Gateway API server",
			},
		}
	} else if g.typeFlag == "botapi" || g.typeFlag == "" {
		info = openapi.Info{
			Title:       "Telegram Bot API",
			Description: `The Bot API is an HTTP-based interface created for developers keen on building bots for Telegram.\nTo learn how to create and set up a bot, please consult [Introduction to Bots](https://core.telegram.org/bots) and [Bot FAQ](https://core.telegram.org/bots/faq).`,
			Version:     g.version,
		}
		servers = []openapi.Server{
			{
				URL:         "https://api.telegram.org/bot{token}/",
				Description: "Production Telegram Bot API server",
				Variables: openapi.Variables{
					Token: &openapi.TokenVariable{
						Description: "Bot token provided by BotFather. It is used to authenticate requests to the Telegram Bot API.",
						Default:     "123456789:ABCdefGHIjklMNOpqrSTUvwxYZ",
					},
				},
			},
			{
				URL:         "https://api.telegram.org/beta/bot{token}/",
				Description: "Beta Telegram Bot API server",
				Variables: openapi.Variables{
					Token: &openapi.TokenVariable{
						Description: "Bot token provided by BotFather. It is used to authenticate requests to the Telegram Bot API.",
						Default:     "123456789:ABCdefGHIjklMNOpqrSTUvwxYZ",
					},
				},
			},
		}
	} else {
		return nil, fmt.Errorf("unknown type: %s", g.typeFlag)
	}

	openAPI := &openapi.OpenAPI{
		OpenAPI: "3.1.0",
		Info:    info,
		Servers: servers,
		Paths:   make(map[string]openapi.Path),
		Components: openapi.Components{
			Schemas: make(map[string]openapi.Schema),
			SecuritySchemes: func() map[string]openapi.SecurityScheme {
				if g.typeFlag == "gateway" {
					return map[string]openapi.SecurityScheme{
						"access_token": {
							Type:         "http",
							Scheme:       "bearer",
							BearerFormat: "JWT",
							Description:  "Access token obtained in the Telegram Gateway account settings.",
						},
					}
				}
				return nil
			}(),
		},
	}

	// Add global security requirement for gateway
	if g.typeFlag == "gateway" {
		openAPI.Security = []map[string][]string{
			{"access_token": {}},
		}
	}

	unionTypes := g.detectUnionTypes()
	g.log.Debug("detected union types", zap.Int("count", len(unionTypes)))
	for name, variants := range unionTypes {
		g.log.Debug("union type", zap.String("name", name), zap.Strings("variants", variants))
	}

	for _, t := range g.types {
		g.log.Debug("processing type", zap.String("name", t.Name))
		if variants, ok := unionTypes[t.Name]; ok {
			schema := openapi.Schema{
				OneOf:       []openapi.Property{},
				Type:        "object",
				Description: t.Description,
			}
			for _, v := range variants {
				schema.OneOf = append(schema.OneOf, openapi.Property{Ref: fmt.Sprintf("#/components/schemas/%s", v)})
			}
			openAPI.Components.Schemas[t.Name] = schema
			continue
		}

		schema := openapi.Schema{
			Type:        "object",
			Properties:  make(map[string]openapi.Property),
			Description: t.Description,
		}

		for _, field := range t.Fields {
			property := g.convertDataTypeToProperty(g.convertStringSliceToDataType(field.Type))
			property.Description = field.Description
			schema.Properties[field.Name] = property
			if field.Required {
				schema.Required = append(schema.Required, field.Name)
			}
		}

		openAPI.Components.Schemas[t.Name] = schema
	}

	for _, m := range g.methods {
		g.log.Debug("processing method", zap.String("name", m.Name))
		properties := make(map[string]openapi.Property)
		required := []string{}
		for _, param := range m.Parameters {
			properties[param.Name] = g.convertDataTypeToProperty(param.Type)
			if param.Required {
				required = append(required, param.Name)
			}
		}

		pathItem := openapi.Path{
			Post: openapi.Operation{
				Summary:     m.Name,
				Description: m.Description,
				OperationID: m.Name,
				RequestBody: openapi.RequestBody{
					Content: openapi.MediaType{
						Applicationjson: openapi.Applicationjson{
							Schema: openapi.Schema{
								Type:       "object",
								Properties: properties,
								Required:   required,
							},
						},
					},
					Required: true,
				},
				Responses: map[string]openapi.Response{
					"200": {
						Description: "Successful response",
						Content: &openapi.MediaType{
							Applicationjson: openapi.Applicationjson{
								Schema: openapi.Schema{
									Type: "object",
									Properties: map[string]openapi.Property{
										"ok": {
											Type:        "boolean",
											Description: "Request success indicator",
										},
										"result": g.convertMethodReturnType(m.ReturnType),
									},
									Required: []string{"ok", "result"},
								},
							},
						},
					},
				},
			},
		}
		openAPI.Paths["/"+m.Name] = pathItem
	}

	g.log.Debug("OpenAPI generation complete")
	return openAPI, nil
}

func (g *Generator) Save(openAPI *openapi.OpenAPI, outputPath string) error {
	g.log.Debug("marshaling OpenAPI to JSON")
	data, err := json.MarshalIndent(openAPI, "", "    ")
	if err != nil {
		g.log.Error("error marshaling JSON", zap.Error(err))
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Determine file path and directory
	path := outputPath
	isDir := false
	if path == "" || path == "." || strings.HasSuffix(path, "/") {
		isDir = true
	}
	if !isDir {
		// If path exists and is a directory, treat as directory
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			isDir = true
		}
	}

	if isDir {
		// Use default file name in the directory
		path = strings.TrimRight(path, "/")
		if path == "" {
			path = "."
		}
		path = path + "/openapi-v%v.json"
	}

	if strings.Contains(path, "%v") {
		path = fmt.Sprintf(path, g.version)
	}
	path = strings.ReplaceAll(strings.ReplaceAll(path, " ", ""), ",", "-")

	dir := "."
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		dir = path[:idx]
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		g.log.Error("failed to create output directory", zap.Error(err), zap.String("dir", dir))
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	g.log.Debug("writing OpenAPI file", zap.String("path", path))
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		g.log.Error("error writing file", zap.Error(err), zap.String("path", path))
		return fmt.Errorf("error write file: %w", err)
	}

	g.log.Info("saved openapi file", zap.String("path", path))
	return nil
}

func (g *Generator) detectUnionTypes() map[string][]string {
	unions := make(map[string][]string)
	for _, t := range g.types {
		desc := t.Description

		if len(t.Fields) == 0 && (strings.Contains(desc, "this object represents") || strings.Contains(desc, "this object describes") || strings.Contains(desc, "should be one of") || strings.Contains(desc, "can be one of")) {

			lines := strings.Split(desc, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if after, ok := strings.CutPrefix(line, "-"); ok {
					candidate := strings.TrimSpace(after)
					if candidate != "" {
						unions[t.Name] = append(unions[t.Name], candidate)
					}
				}
			}
		}
	}
	return unions
}

func (g *Generator) convertDataTypeToProperty(dt telegram.DataType) openapi.Property {
	if dt.IsArray {
		var innerProperty openapi.Property
		if len(dt.Types) > 1 {
			innerProperty = openapi.Property{
				OneOf: g.convertTypesToProperties(dt.Types),
			}
		} else if strings.HasPrefix(g.convertType(dt.Types[0]), "#/components/schemas/") {
			innerProperty = openapi.Property{
				Ref: g.convertType(dt.Types[0]),
			}
		} else {
			innerProperty = openapi.Property{
				Type: g.convertType(dt.Types[0]),
			}
		}

		result := openapi.Property{
			Type:  "array",
			Items: &innerProperty,
		}

		for i := 1; i < dt.ArrayDepth; i++ {
			result = openapi.Property{
				Type: "array",
				Items: &openapi.Property{
					Type:  "array",
					Items: result.Items,
				},
			}
		}
		return result
	}

	if len(dt.Types) > 1 {
		return openapi.Property{
			OneOf: g.convertTypesToProperties(dt.Types),
		}
	}

	if strings.HasPrefix(g.convertType(dt.Types[0]), "#/components/schemas/") {
		return openapi.Property{
			Ref: g.convertType(dt.Types[0]),
		}
	} else {
		return openapi.Property{
			Type: g.convertType(dt.Types[0]),
		}
	}
}

func (g *Generator) convertTypesToProperties(types []string) []openapi.Property {
	properties := make([]openapi.Property, 0, len(types))
	for _, t := range types {
		if strings.HasPrefix(g.convertType(t), "#/components/schemas/") {
			properties = append(properties, openapi.Property{
				Ref: g.convertType(t),
			})
		} else {
			properties = append(properties, openapi.Property{
				Type: g.convertType(t),
			})
		}
	}
	return properties
}

func (g *Generator) convertType(t string) string {
	switch t {
	case "Integer", "Int":
		return "integer"
	case "Float", "Double":
		return "number"
	case "Boolean", "Bool", "True", "False":
		return "boolean"
	case "String":
		return "string"
	default:
		return fmt.Sprintf("#/components/schemas/%s", t)
	}
}

func (g *Generator) convertStringSliceToDataType(types []string) telegram.DataType {
	isArray := false
	arrayDepth := 0
	cleanTypes := make([]string, 0, len(types))

	for _, t := range types {
		currentType := t
		for strings.HasPrefix(currentType, "Array of ") {
			isArray = true
			arrayDepth++
			currentType = strings.TrimPrefix(currentType, "Array of ")
		}
		cleanTypes = append(cleanTypes, currentType)
	}

	return telegram.DataType{
		Types:      cleanTypes,
		IsArray:    isArray,
		ArrayDepth: arrayDepth,
	}
}

func (g *Generator) convertMethodReturnType(returnType telegram.ReturnType) openapi.Property {
	if returnType.IsArray {
		return g.convertDataTypeToProperty(telegram.DataType{
			Types:      []string{returnType.Name},
			IsArray:    true,
			ArrayDepth: 1,
		})
	}

	switch returnType.Name {
	case "":
		return openapi.Property{Type: "boolean"}
	case "integer":
		return openapi.Property{Type: "integer"}
	case "boolean":
		return openapi.Property{Type: "boolean"}
	case "String":
		return openapi.Property{Type: "string"}
	case "True", "False":
		return openapi.Property{Type: "boolean"}
	default:
		return openapi.Property{
			Ref: fmt.Sprintf("#/components/schemas/%s", returnType.Name),
		}
	}
}
