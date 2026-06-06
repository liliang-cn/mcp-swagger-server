package mcp

import (
	"encoding/json"
	"testing"
)

// asJSON round-trips an arbitrary schema value into a generic map for assertions.
func asJSON(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}
	return m
}

// bodySchemaOf parses a spec, finds the POST /profiles operation and returns
// the generated input schema for assertions.
func bodySchemaOf(t *testing.T, specData string) map[string]interface{} {
	t.Helper()
	swagger, err := ParseSwaggerSpec([]byte(specData))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	pathItem, ok := swagger.Paths.Paths["/profiles"]
	if !ok || pathItem.Post == nil {
		t.Fatalf("POST /profiles not found after parsing; paths=%v", swagger.Paths.Paths)
	}

	s := &SwaggerMCPServer{swagger: swagger}
	return asJSON(t, s.buildParametersSchema(pathItem.Post.Parameters))
}

// requireBodyFields asserts the generated tool input schema exposes the
// request body fields id/node/delay/keep_masked with id required.
func requireBodyFields(t *testing.T, schema map[string]interface{}) {
	t.Helper()

	props, _ := schema["properties"].(map[string]interface{})
	body, _ := props["body"].(map[string]interface{})
	if body == nil {
		t.Fatalf("schema has no body property: %v", schema)
	}

	bodyProps, _ := body["properties"].(map[string]interface{})
	for _, field := range []string{"id", "node", "delay", "keep_masked"} {
		if _, ok := bodyProps[field]; !ok {
			t.Errorf("body schema missing field %q: %v", field, body)
		}
	}

	required, _ := body["required"].([]interface{})
	found := false
	for _, r := range required {
		if r == "id" {
			found = true
		}
	}
	if !found {
		t.Errorf("body schema does not mark 'id' as required: %v", body)
	}
}

// TestParseSwaggerSpec_OpenAPI3RequestBody verifies that OpenAPI 3.0 specs
// with requestBody produce a full input schema (previously the requestBody
// was silently dropped, yielding zero tool parameters).
func TestParseSwaggerSpec_OpenAPI3RequestBody(t *testing.T) {
	specData := `{
	  "openapi": "3.0.3",
	  "info": {"title": "Repro", "version": "1.0"},
	  "paths": {
	    "/profiles": {
	      "post": {
	        "operationId": "create_profile",
	        "summary": "Create HA profile",
	        "requestBody": {
	          "required": true,
	          "content": {
	            "application/json": {
	              "schema": {
	                "type": "object",
	                "required": ["id"],
	                "properties": {
	                  "id": {"type": "string", "description": "Profile ID"},
	                  "node": {"type": "string"},
	                  "delay": {"type": "integer"},
	                  "keep_masked": {"type": "boolean"}
	                }
	              }
	            }
	          }
	        },
	        "responses": {"201": {"description": "Created"}}
	      }
	    }
	  }
	}`
	requireBodyFields(t, bodySchemaOf(t, specData))
}

// TestParseSwaggerSpec_OpenAPI31Utoipa verifies OpenAPI 3.1 specs as emitted
// by utoipa (type arrays with "null", components refs) are supported.
func TestParseSwaggerSpec_OpenAPI31Utoipa(t *testing.T) {
	specData := `{
	  "openapi": "3.1.0",
	  "info": {"title": "drbd-ha", "version": "1.0"},
	  "paths": {
	    "/profiles": {
	      "post": {
	        "operationId": "create_profile",
	        "summary": "Create HA profile",
	        "requestBody": {
	          "required": true,
	          "content": {
	            "application/json": {
	              "schema": {"$ref": "#/components/schemas/CreateProfile"}
	            }
	          }
	        },
	        "responses": {"201": {"description": "Created"}}
	      }
	    }
	  },
	  "components": {
	    "schemas": {
	      "CreateProfile": {
	        "type": "object",
	        "required": ["id"],
	        "properties": {
	          "id": {"type": "string"},
	          "node": {"type": ["string", "null"]},
	          "delay": {"type": ["integer", "null"]},
	          "keep_masked": {"type": "boolean"}
	        }
	      }
	    }
	  }
	}`
	requireBodyFields(t, bodySchemaOf(t, specData))
}

// TestParseSwaggerSpec_Swagger2RefExpansion verifies that $ref body schemas
// in Swagger 2.0 specs are expanded instead of degrading to a bare object.
func TestParseSwaggerSpec_Swagger2RefExpansion(t *testing.T) {
	specData := `{
	  "swagger": "2.0",
	  "info": {"title": "Repro2", "version": "1.0"},
	  "paths": {
	    "/profiles": {
	      "post": {
	        "operationId": "create_profile",
	        "summary": "Create HA profile",
	        "parameters": [
	          {"name": "body", "in": "body", "required": true,
	           "schema": {"$ref": "#/definitions/CreateProfile"}}
	        ],
	        "responses": {"201": {"description": "Created"}}
	      }
	    }
	  },
	  "definitions": {
	    "CreateProfile": {
	      "type": "object",
	      "required": ["id"],
	      "properties": {
	        "id": {"type": "string", "description": "Profile ID"},
	        "node": {"type": "string"},
	        "delay": {"type": "integer"},
	        "keep_masked": {"type": "boolean"}
	      }
	    }
	  }
	}`
	requireBodyFields(t, bodySchemaOf(t, specData))
}

// TestParseSwaggerSpec_NestedBodySchema verifies nested objects and arrays in
// body schemas survive into the generated input schema.
func TestParseSwaggerSpec_NestedBodySchema(t *testing.T) {
	specData := `{
	  "swagger": "2.0",
	  "info": {"title": "Nested", "version": "1.0"},
	  "paths": {
	    "/profiles": {
	      "post": {
	        "operationId": "create_profile",
	        "parameters": [
	          {"name": "body", "in": "body", "required": true,
	           "schema": {
	             "type": "object",
	             "properties": {
	               "config": {
	                 "type": "object",
	                 "properties": {"timeout": {"type": "integer"}}
	               },
	               "nodes": {
	                 "type": "array",
	                 "items": {"type": "string"}
	               }
	             }
	           }}
	        ],
	        "responses": {"201": {"description": "Created"}}
	      }
	    }
	  }
	}`
	schema := bodySchemaOf(t, specData)

	props, _ := schema["properties"].(map[string]interface{})
	body, _ := props["body"].(map[string]interface{})
	bodyProps, _ := body["properties"].(map[string]interface{})

	config, _ := bodyProps["config"].(map[string]interface{})
	configProps, _ := config["properties"].(map[string]interface{})
	if _, ok := configProps["timeout"]; !ok {
		t.Errorf("nested object field config.timeout lost: %v", body)
	}

	nodes, _ := bodyProps["nodes"].(map[string]interface{})
	items, _ := nodes["items"].(map[string]interface{})
	if items["type"] != "string" {
		t.Errorf("array items schema lost: %v", body)
	}
}
