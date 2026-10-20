package plugin

import (
	"encoding/json"
	"strings"
	"testing"
)

// write mimics the host writing data into the plugin's shared bytes
func write(p *Plugin, data string) {
	p.ClearSharedBytes(uint32(len(data)))
	copy(p.sharedBytes, data)
}

func TestInfo(t *testing.T) {
	var p Plugin
	if size := p.Info("1.2.3"); size != uint32(len(p.sharedBytes)) {
		t.Fatalf("Info() = %d, want %d", size, len(p.sharedBytes))
	}
	var info struct {
		Name            string `json:"name"`
		Version         string `json:"version"`
		ConfigKey       string `json:"configKey"`
		HelpURL         string `json:"helpUrl"`
		ConfigSchemaURL string `json:"configSchemaUrl"`
		UpdateURL       string `json:"updateUrl"`
	}
	if err := json.Unmarshal(p.takeFromSharedBytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Name != "dprint-plugin-terraform" ||
		info.Version != "1.2.3" ||
		info.ConfigKey != "terraform" ||
		info.HelpURL != "https://github.com/dprint/dprint-plugin-terraform" ||
		info.ConfigSchemaURL != "https://plugins.dprint.dev/dprint/dprint-plugin-terraform/1.2.3/schema.json" ||
		info.UpdateURL != "https://plugins.dprint.dev/dprint/dprint-plugin-terraform/latest.json" {
		t.Fatalf("unexpected plugin info: %#v", info)
	}
}

func TestConfigFileMatching(t *testing.T) {
	var p Plugin
	p.ConfigFileMatching(1)
	var matching struct {
		FileExtensions []string `json:"fileExtensions"`
		FileNames      []string `json:"fileNames"`
		Additive       bool     `json:"additive"`
	}
	if err := json.Unmarshal(p.takeFromSharedBytes(), &matching); err != nil {
		t.Fatal(err)
	}
	if strings.Join(matching.FileExtensions, ",") != "tf,tofu,tfvars,hcl" || len(matching.FileNames) != 0 || matching.Additive {
		t.Fatalf("unexpected file matching: %#v", matching)
	}
}

func TestConfigsByID(t *testing.T) {
	var p Plugin
	write(&p, `{"plugin":{},"global":{"indentWidth":4,"newLineKind":"crlf"}}`)
	p.RegisterConfig(1)
	write(&p, `{"plugin":{"indentWidth":4,"unknown":null},"global":{}}`)
	p.RegisterConfig(2)

	p.ResolvedConfig(1)
	if resolved := string(p.takeFromSharedBytes()); resolved != "{}" {
		t.Fatalf("unexpected resolved config: %s", resolved)
	}

	// the global config doesn't apply, so it gets no diagnostics
	p.ConfigDiagnostics(1)
	if diagnostics := string(p.takeFromSharedBytes()); diagnostics != "[]" {
		t.Fatalf("unexpected diagnostics for config 1: %s", diagnostics)
	}
	p.ConfigDiagnostics(2)
	expected := `[{"propertyName":"indentWidth","message":"Unknown property in configuration"},` +
		`{"propertyName":"unknown","message":"Unknown property in configuration"}]`
	if diagnostics := string(p.takeFromSharedBytes()); diagnostics != expected {
		t.Fatalf("unexpected diagnostics for config 2: %s", diagnostics)
	}

	p.ReleaseConfig(1)
	if _, ok := p.configs[1]; ok {
		t.Fatal("config 1 was not released")
	}
}

func TestFormat(t *testing.T) {
	var p Plugin
	write(&p, `{"plugin":{},"global":{}}`)
	p.RegisterConfig(1)

	write(&p, `C:\repo\main.tf`)
	p.SetFilePath()
	write(&p, "resource aws_instance web {\nami = \"${var.ami}\"\n}\n")
	if result := p.Format(1); result != formatChanged {
		t.Fatalf("Format() = %d, want %d", result, formatChanged)
	}
	p.FormattedText()
	if formatted := string(p.takeFromSharedBytes()); formatted != "resource \"aws_instance\" \"web\" {\n  ami = var.ami\n}\n" {
		t.Fatalf("unexpected formatted text: %q", formatted)
	}
	p.FormattedText()
	if formatted := p.takeFromSharedBytes(); formatted != nil {
		t.Fatalf("FormattedText() did not clear the result: %q", formatted)
	}

	// generic HCL files don't get the terraform rewrites
	write(&p, "/repo/terragrunt.hcl")
	p.SetFilePath()
	write(&p, "inputs = {\nname = \"${local.name}\"\n}\n")
	if result := p.Format(1); result != formatChanged {
		t.Fatalf("Format() = %d, want %d", result, formatChanged)
	}
	p.FormattedText()
	if formatted := string(p.takeFromSharedBytes()); formatted != "inputs = {\n  name = \"${local.name}\"\n}\n" {
		t.Fatalf("unexpected formatted text: %q", formatted)
	}

	write(&p, "/repo/main.tf")
	p.SetFilePath()
	write(&p, "a = 1\n")
	if result := p.Format(1); result != formatNoChange {
		t.Fatalf("Format() = %d, want %d", result, formatNoChange)
	}

	write(&p, "/repo/main.tf")
	p.SetFilePath()
	write(&p, "a = \n")
	if result := p.Format(1); result != formatError {
		t.Fatalf("Format() = %d, want %d", result, formatError)
	}
	p.ErrorText()
	if errorText := string(p.takeFromSharedBytes()); errorText != "/repo/main.tf:1:5: Invalid expression; Expected the start of an expression, but found an invalid expression token." {
		t.Fatalf("unexpected error text: %q", errorText)
	}
}

func TestFormatRange(t *testing.T) {
	var p Plugin
	write(&p, `{"plugin":{},"global":{}}`)
	p.RegisterConfig(1)

	write(&p, `{"some":"override"}`)
	p.SetOverrideConfig()
	write(&p, "/repo/main.tf")
	p.SetFilePath()
	write(&p, "a=1\n")
	if result := p.FormatRange(1, 0, 2); result != formatNoChange {
		t.Fatalf("FormatRange() = %d, want %d", result, formatNoChange)
	}
	if p.sharedBytes != nil || p.filePath != "" || p.overrideConfig != nil {
		t.Fatal("FormatRange() did not clear the state of the format")
	}
}

func TestCheckConfigUpdates(t *testing.T) {
	var p Plugin
	write(&p, `{"oldVersion":"0.0.1","config":{}}`)
	p.CheckConfigUpdates()
	if result := string(p.takeFromSharedBytes()); result != `{"kind":"ok","data":[]}` {
		t.Fatalf("unexpected result: %s", result)
	}
}
