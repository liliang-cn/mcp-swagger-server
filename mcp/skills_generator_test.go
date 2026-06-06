package mcp

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Per https://agentskills.io/specification the frontmatter `name`:
//   - 1-64 chars, lowercase a-z, 0-9 and hyphens only
//   - must not start/end with a hyphen, no consecutive hyphens
//   - must match the parent directory name
var validSkillName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

const skillsTestSwagger = `{
  "swagger": "2.0",
  "info": {"title": "Test API", "version": "1.0.0"},
  "paths": {
    "/pets": {
      "get": {
        "operationId": "listPets",
        "summary": "List all pets",
        "tags": ["pets"],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "operationId": "createPet",
        "summary": "Create a pet",
        "tags": ["pets"],
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/pets/{petId}": {
      "get": {
        "operationId": "getPetById",
        "summary": "Get a pet by ID",
        "tags": ["pets"],
        "responses": {"200": {"description": "OK"}}
      },
      "delete": {
        "operationId": "deletePet",
        "summary": "Delete a pet",
        "tags": ["pets"],
        "responses": {"204": {"description": "Deleted"}}
      }
    },
    "/orders": {
      "post": {
        "operationId": "createOrder",
        "summary": "Create an order",
        "tags": ["Order Management!"],
        "responses": {"201": {"description": "Created"}}
      }
    }
  }
}`

func parseFrontmatter(t *testing.T, content string) map[string]string {
	t.Helper()
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		t.Fatalf("SKILL.md has no YAML frontmatter delimited by ---")
	}
	// The frontmatter must be valid YAML — agents parse it with a real
	// YAML parser, so unquoted values containing ': ' must not slip through.
	fields := map[string]string{}
	if err := yaml.Unmarshal([]byte(parts[1]), &fields); err != nil {
		t.Fatalf("SKILL.md frontmatter is not valid YAML: %v\n%s", err, parts[1])
	}
	return fields
}

// TestGenerateSkills_AgentSkillsSpec verifies that generated skills conform to
// the Agent Skills specification (https://agentskills.io/specification).
func TestGenerateSkills_AgentSkillsSpec(t *testing.T) {
	swagger, err := ParseSwaggerSpec([]byte(skillsTestSwagger))
	if err != nil {
		t.Fatalf("failed to parse swagger: %v", err)
	}

	outDir := t.TempDir()
	generator := NewSkillsGenerator(swagger, "http://localhost:4538/v2", outDir)
	if err := generator.Generate(); err != nil {
		t.Fatalf("failed to generate skills: %v", err)
	}

	skillFiles, err := filepath.Glob(filepath.Join(outDir, "*", "SKILL.md"))
	if err != nil || len(skillFiles) == 0 {
		t.Fatalf("no SKILL.md files generated (err=%v)", err)
	}

	for _, skillFile := range skillFiles {
		dirName := filepath.Base(filepath.Dir(skillFile))
		content, err := os.ReadFile(skillFile)
		if err != nil {
			t.Fatalf("failed to read %s: %v", skillFile, err)
		}
		fields := parseFrontmatter(t, string(content))

		name := fields["name"]
		if name == "" {
			t.Errorf("%s: missing required frontmatter field 'name'", skillFile)
			continue
		}
		if len(name) > 64 {
			t.Errorf("%s: name %q exceeds 64 characters", skillFile, name)
		}
		if !validSkillName.MatchString(name) {
			t.Errorf("%s: name %q violates spec (lowercase a-z0-9 + single hyphens only)", skillFile, name)
		}
		if name != dirName {
			t.Errorf("%s: name %q does not match parent directory name %q", skillFile, name, dirName)
		}

		desc := fields["description"]
		if desc == "" {
			t.Errorf("%s: missing required frontmatter field 'description'", skillFile)
		}
		if len(desc) > 1024 {
			t.Errorf("%s: description exceeds 1024 characters", skillFile)
		}
		// The spec says the description should say what the skill does AND
		// when to use it, with trigger keywords.
		if !strings.Contains(strings.ToLower(desc), "use when") &&
			!strings.Contains(strings.ToLower(desc), "use this skill") {
			t.Errorf("%s: description %q lacks 'when to use' guidance", skillFile, desc)
		}
	}

	// Directory names themselves must be valid skill names too
	// (sanitizeName must not emit leading/trailing/consecutive hyphens).
	for _, skillFile := range skillFiles {
		dirName := filepath.Base(filepath.Dir(skillFile))
		if !validSkillName.MatchString(dirName) {
			t.Errorf("skill directory %q violates spec naming rules", dirName)
		}
	}
}
