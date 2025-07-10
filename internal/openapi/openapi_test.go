package openapi

import (
	"encoding/json"
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
