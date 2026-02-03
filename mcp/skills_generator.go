package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
)

// SkillsGenerator converts Swagger specifications to Agent Skills format
type SkillsGenerator struct {
	swagger    *spec.Swagger
	apiBaseURL string
	outputDir  string
}

// NewSkillsGenerator creates a new Skills generator
func NewSkillsGenerator(swagger *spec.Swagger, apiBaseURL, outputDir string) *SkillsGenerator {
	return &SkillsGenerator{
		swagger:    swagger,
		apiBaseURL: apiBaseURL,
		outputDir:  outputDir,
	}
}

// Generate generates all Skills from the Swagger specification
func (sg *SkillsGenerator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(sg.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Group endpoints by tag for organization
	tagGroups := sg.groupByTag()

	// Generate a skill for each tag group
	for tag, operations := range tagGroups {
		if err := sg.generateSkillForTag(tag, operations); err != nil {
			return fmt.Errorf("failed to generate skill for tag %s: %w", tag, err)
		}
	}

	// Generate index file with all skills metadata
	if err := sg.generateSkillsIndex(tagGroups); err != nil {
		return fmt.Errorf("failed to generate skills index: %w", err)
	}

	return nil
}

// Operation represents an API operation for skill generation
type Operation struct {
	Method   string
	Path     string
	Spec     *spec.Operation
	Tag      string
	ToolName string
}

// groupByTag groups operations by their tags
func (sg *SkillsGenerator) groupByTag() map[string][]Operation {
	groups := make(map[string][]Operation)

	for path, pathItem := range sg.swagger.Paths.Paths {
		operations := []struct {
			method string
			op     *spec.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
		}

		for _, op := range operations {
			if op.op == nil {
				continue
			}

			// Get tags - default to "default" if no tags
			tags := op.op.Tags
			if len(tags) == 0 {
				tags = []string{"default"}
			}

			toolName := GenerateToolName(op.method, path, op.op)

			for _, tag := range tags {
				// Clean tag name for use as directory name
				cleanTag := sanitizeName(tag)
				groups[cleanTag] = append(groups[cleanTag], Operation{
					Method:   op.method,
					Path:     path,
					Spec:     op.op,
					Tag:      tag,
					ToolName: toolName,
				})
			}
		}
	}

	return groups
}

// generateSkillForTag creates a skill directory and files for a tag group
func (sg *SkillsGenerator) generateSkillForTag(tag string, operations []Operation) error {
	// Create skill directory
	skillDir := filepath.Join(sg.outputDir, tag)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	// Generate SKILL.md
	skillMD := sg.generateSkillMD(tag, operations)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		return err
	}

	// Generate reference.md with detailed API documentation
	referenceMD := sg.generateReferenceMD(tag, operations)
	if err := os.WriteFile(filepath.Join(skillDir, "reference.md"), []byte(referenceMD), 0644); err != nil {
		return err
	}

	return nil
}

// generateSkillMD creates the SKILL.md content
func (sg *SkillsGenerator) generateSkillMD(tag string, operations []Operation) string {
	name := toTitleCase(tag)
	description := sg.generateDescription(tag, operations)

	var sb strings.Builder

	// YAML frontmatter (required)
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", name))
	sb.WriteString(fmt.Sprintf("description: %s\n", description))
	sb.WriteString("---\n\n")

	// Skill introduction
	sb.WriteString(fmt.Sprintf("# %s API Skill\n\n", name))
	sb.WriteString(fmt.Sprintf("This skill provides tools for interacting with the %s API.\n\n", tag))

	// When to use this skill
	sb.WriteString("## When to use this skill\n\n")
	sb.WriteString("Use this skill when you need to:\n")
	for _, op := range operations {
		if op.Spec.ID != "" {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", op.Spec.ID, getSummary(op.Spec)))
		}
	}
	sb.WriteString("\n")

	// Available tools
	sb.WriteString("## Available Tools\n\n")
	for _, op := range operations {
		sb.WriteString(fmt.Sprintf("### %s\n", op.ToolName))
		sb.WriteString(fmt.Sprintf("**%s %s**\n\n", op.Method, op.Path))
		if op.Spec.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", op.Spec.Description))
		} else if op.Spec.Summary != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", op.Spec.Summary))
		}
	}

	// Authentication
	if sg.apiBaseURL != "" {
		sb.WriteString("## Configuration\n\n")
		sb.WriteString(fmt.Sprintf("- **Base URL**: `%s`\n", sg.apiBaseURL))
		sb.WriteString("\n")
	}

	// Reference to additional documentation
	sb.WriteString("## Additional Documentation\n\n")
	sb.WriteString("See [reference.md](reference.md) for detailed API documentation including:\n")
	sb.WriteString("- Request/response schemas\n")
	sb.WriteString("- Parameter descriptions\n")
	sb.WriteString("- Error handling\n")
	sb.WriteString("- Usage examples\n")

	return sb.String()
}

// generateReferenceMD creates detailed reference documentation
func (sg *SkillsGenerator) generateReferenceMD(tag string, operations []Operation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s API Reference\n\n", toTitleCase(tag)))

	for _, op := range operations {
		sb.WriteString(fmt.Sprintf("## %s\n\n", op.ToolName))
		sb.WriteString(fmt.Sprintf("**Method**: %s\n\n", op.Method))
		sb.WriteString(fmt.Sprintf("**Path**: `%s`\n\n", op.Path))

		if op.Spec.Summary != "" {
			sb.WriteString(fmt.Sprintf("**Summary**: %s\n\n", op.Spec.Summary))
		}

		if op.Spec.Description != "" {
			sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", op.Spec.Description))
		}

		// Parameters
		if len(op.Spec.Parameters) > 0 {
			sb.WriteString("### Parameters\n\n")
			sb.WriteString("| Name | In | Type | Required | Description |\n")
			sb.WriteString("|------|-----|------|----------|-------------|\n")
			for _, param := range op.Spec.Parameters {
				required := "No"
				if param.Required {
					required = "Yes"
				}
				paramType := param.Type
				if paramType == "" && param.Schema != nil && len(param.Schema.Type) > 0 {
					paramType = param.Schema.Type[0]
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					param.Name, param.In, paramType, required, getParamDescription(param)))
			}
			sb.WriteString("\n")
		}

		// Responses
		if op.Spec.Responses != nil && len(op.Spec.Responses.StatusCodeResponses) > 0 {
			sb.WriteString("### Responses\n\n")
			for code, response := range op.Spec.Responses.StatusCodeResponses {
				sb.WriteString(fmt.Sprintf("#### %s\n", code))
				if response.Description != "" {
					sb.WriteString(fmt.Sprintf("%s\n\n", response.Description))
				}
				if response.Schema != nil {
					sb.WriteString("**Schema**:\n```json\n")
					sb.WriteString(schemaToExample(response.Schema))
					sb.WriteString("\n```\n\n")
				}
			}
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// generateSkillsIndex creates an index of all skills
func (sg *SkillsGenerator) generateSkillsIndex(tagGroups map[string][]Operation) error {
	var sb strings.Builder

	sb.WriteString("# API Skills Index\n\n")
	sb.WriteString("This directory contains Agent Skills generated from the OpenAPI/Swagger specification.\n\n")
	sb.WriteString(fmt.Sprintf("**API Base URL**: %s\n\n", sg.apiBaseURL))
	sb.WriteString("## Available Skills\n\n")

	for tag, operations := range tagGroups {
		sb.WriteString(fmt.Sprintf("### [%s](./%s/SKILL.md)\n\n", toTitleCase(tag), tag))
		sb.WriteString(fmt.Sprintf("%s\n\n", sg.generateDescription(tag, operations)))
		sb.WriteString(fmt.Sprintf("**Operations**: %d\n\n", len(operations)))
	}

	return os.WriteFile(filepath.Join(sg.outputDir, "INDEX.md"), []byte(sb.String()), 0644)
}

// generateDescription creates a description for a tag group
func (sg *SkillsGenerator) generateDescription(tag string, operations []Operation) string {
	if len(operations) == 0 {
		return fmt.Sprintf("API operations for %s", tag)
	}

	// Create description from operations
	var methods []string
	for _, op := range operations {
		if op.Spec.Summary != "" {
			methods = append(methods, op.Spec.Summary)
		} else if op.Spec.ID != "" {
			methods = append(methods, op.Spec.ID)
		}
	}

	if len(methods) > 0 && len(methods) <= 3 {
		return fmt.Sprintf("Provides %s", joinWithComma(methods))
	}

	return fmt.Sprintf("API operations for %s (%d endpoints)", tag, len(operations))
}

// Helper functions

func getSummary(op *spec.Operation) string {
	if op.Summary != "" {
		return op.Summary
	}
	if op.Description != "" {
		// Truncate description if too long
		if len(op.Description) > 80 {
			return op.Description[:77] + "..."
		}
		return op.Description
	}
	if op.ID != "" {
		return op.ID
	}
	return "API operation"
}

func getParamDescription(param spec.Parameter) string {
	if param.Description != "" {
		return strings.ReplaceAll(param.Description, "|", "\\|")
	}
	return "-"
}

func schemaToExample(schema *spec.Schema) string {
	if schema == nil {
		return "{}"
	}
	// Simple schema to JSON example conversion
	if len(schema.Type) > 0 {
		switch schema.Type[0] {
		case "object":
			return "{\n  \"key\": \"value\"\n}"
		case "array":
			return "[]"
		case "string":
			return "\"string\""
		case "number", "integer":
			return "0"
		case "boolean":
			return "true"
		}
	}
	return "{}"
}
