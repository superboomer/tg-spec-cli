package openapi

type OpenAPI struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers"`
	Paths      map[string]Path       `json:"paths"`
	Components Components            `json:"components,omitempty"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type Info struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

type Server struct {
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Variables   Variables `json:"variables,omitempty"`
}

type Variables struct {
	Token       *TokenVariable `json:"token,omitempty"`
	AccessToken *TokenVariable `json:"access_token,omitempty"`
}

type TokenVariable struct {
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

type Path struct {
	Post Operation `json:"post"`
}

type Operation struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	OperationID string              `json:"operationId"`
	RequestBody RequestBody         `json:"requestBody"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Content  MediaType `json:"content"`
	Required bool      `json:"required"`
}

type MediaType struct {
	Applicationjson Applicationjson `json:"application/json,omitempty"`
}

type Applicationjson struct {
	Schema Schema `json:"schema,omitempty"`
}

type Schema struct {
	Type        string              `json:"type,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
	Description string              `json:"description,omitempty"`
	OneOf       []Property          `json:"oneOf,omitempty"`
}

type Property struct {
	Ref         string     `json:"$ref,omitempty"`
	Type        string     `json:"type,omitempty"`
	Items       *Property  `json:"items,omitempty"`
	OneOf       []Property `json:"oneOf,omitempty"`
	Description string     `json:"description,omitempty"`
}

type Response struct {
	Description string     `json:"description"`
	Content     *MediaType `json:"content,omitempty"`
}

type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}
