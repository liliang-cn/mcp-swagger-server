package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// isOpenAPI3 reports whether the (JSON) spec document declares OpenAPI 3.x.
func isOpenAPI3(jsonData []byte) bool {
	var probe struct {
		OpenAPI string `json:"openapi"`
	}
	if err := json.Unmarshal(jsonData, &probe); err != nil {
		return false
	}
	return len(probe.OpenAPI) > 0 && probe.OpenAPI[0] == '3'
}

// convertOpenAPI3ToSwagger2 converts an OpenAPI 3.x JSON document to a
// Swagger 2.0 JSON document so the rest of the pipeline (go-openapi/spec)
// can consume it. OpenAPI 3.1 documents are first normalized to 3.0.
func convertOpenAPI3ToSwagger2(jsonData []byte) ([]byte, error) {
	jsonData, err := normalizeOpenAPI31(jsonData)
	if err != nil {
		return nil, err
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI 3 spec: %w", err)
	}

	// kin-openapi's FromV3 dereferences doc.Components unconditionally
	if doc.Components == nil {
		doc.Components = &openapi3.Components{}
	}

	var v2 *openapi2.T
	v2, err = openapi2conv.FromV3(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OpenAPI 3 to Swagger 2.0: %w", err)
	}

	out, err := json.Marshal(v2)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize converted spec: %w", err)
	}
	return out, nil
}

// normalizeOpenAPI31 rewrites OpenAPI 3.1-only constructs into their 3.0
// equivalents so kin-openapi can load the document:
//   - "openapi": "3.1.x"            -> "3.0.3"
//   - "type": ["T", "null"]         -> "type": "T", "nullable": true
//   - schema-level "examples": [x]  -> "example": x
//   - "const": v                    -> "enum": [v]
func normalizeOpenAPI31(jsonData []byte) ([]byte, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI 3 JSON: %w", err)
	}

	version, _ := doc["openapi"].(string)
	if len(version) < 3 || version[:3] != "3.1" {
		return jsonData, nil // already 3.0.x
	}

	doc["openapi"] = "3.0.3"
	normalizeSchemaNode(doc)

	return json.Marshal(doc)
}

// normalizeSchemaNode walks the document tree and rewrites 3.1 schema
// keywords in place. It is intentionally permissive: it applies the schema
// rewrites wherever the shapes match, which is safe for the keywords below.
func normalizeSchemaNode(node interface{}) {
	switch v := node.(type) {
	case map[string]interface{}:
		// "type": ["string", "null"] -> "type": "string", "nullable": true
		if types, ok := v["type"].([]interface{}); ok {
			var nonNull []interface{}
			nullable := false
			for _, t := range types {
				if t == "null" {
					nullable = true
				} else {
					nonNull = append(nonNull, t)
				}
			}
			if len(nonNull) == 1 {
				v["type"] = nonNull[0]
				if nullable {
					v["nullable"] = true
				}
			}
		}

		// "const": x -> "enum": [x]
		if c, ok := v["const"]; ok {
			v["enum"] = []interface{}{c}
			delete(v, "const")
		}

		// schema-level "examples": [x, ...] -> "example": x
		// (only when it's an array; parameter/media-type "examples" are maps)
		if examples, ok := v["examples"].([]interface{}); ok && len(examples) > 0 {
			if _, hasExample := v["example"]; !hasExample {
				v["example"] = examples[0]
			}
			delete(v, "examples")
		}

		for _, child := range v {
			normalizeSchemaNode(child)
		}
	case []interface{}:
		for _, child := range v {
			normalizeSchemaNode(child)
		}
	}
}
