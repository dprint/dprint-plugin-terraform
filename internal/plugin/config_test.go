package plugin

import (
	"encoding/json"
	"os"
	"testing"
)

func TestResolveConfigDefaults(t *testing.T) {
	config, diagnostics := resolveConfig(nil)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if config != (Configuration{}) {
		t.Fatalf("unexpected config: %#v", config)
	}
}

// the schema has to describe the same (lack of) options as the plugin
func TestSchemaMatchesConfig(t *testing.T) {
	data, err := os.ReadFile("../../deployment/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		ID         string                     `json:"$id"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.ID != "https://plugins.dprint.dev/dprint/dprint-plugin-terraform/0.0.0/schema.json" {
		t.Fatalf("unexpected $id: %s", schema.ID)
	}
	var resolved map[string]any
	if err := json.Unmarshal(marshalConfig(Configuration{}), &resolved); err != nil {
		t.Fatal(err)
	}
	if len(schema.Properties) != len(resolved) {
		t.Fatalf("the schema has %d properties, but the config has %d", len(schema.Properties), len(resolved))
	}
}
