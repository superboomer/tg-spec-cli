package openapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenAPIStructMarshal(t *testing.T) {
	o := OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       "Test API",
			Description: "A test API",
			Version:     "1.0.0",
		},
		Servers: []Server{{
			URL:         "http://localhost",
			Description: "Local server",
		}},
		Paths: map[string]Path{
			"/test": {
				Post: Operation{
					Summary:     "Test operation",
					Description: "A test operation",
					OperationID: "testOp",
					RequestBody: RequestBody{
						Content:  MediaType{},
						Required: true,
					},
					Responses: map[string]Response{},
				},
			},
		},
	}
	_, err := json.Marshal(o)
	if err != nil {
		t.Errorf("OpenAPI struct failed to marshal: %v", err)
	}
}

func TestServerVariablesOmitEmpty(t *testing.T) {
	// A server without variables (e.g. the gateway API) must not emit an empty
	// "variables" object, while a server with a token variable must include it.
	withVars := Server{
		URL: "https://api.telegram.org/bot{token}/",
		Variables: &Variables{
			Token: &TokenVariable{Description: "Bot token", Default: "123:ABC"},
		},
	}
	withoutVars := Server{URL: "https://gatewayapi.telegram.org/"}

	b1, err := json.Marshal(withVars)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b1), `"variables"`) {
		t.Errorf("expected variables to be present, got %s", b1)
	}

	b2, err := json.Marshal(withoutVars)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b2), `"variables"`) {
		t.Errorf("expected variables to be omitted, got %s", b2)
	}
}
