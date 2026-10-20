package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/dprint/dprint-plugin-terraform/internal/tofufmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// FileExtensions are the extensions of the files the plugin formats. dprint
// only matches the last extension of a file, so `.tftest.hcl` and
// `.tofutest.hcl` files come in through `hcl`.
var FileExtensions = []string{"tf", "tofu", "tfvars", "hcl"}

// FormatText formats Terraform and OpenTofu files (`.tf`, `.tofu`, `.tfvars`,
// `.tftest.hcl` and `.tofutest.hcl`) the same way `tofu fmt` does, and other
// HCL files the same way `hclfmt` does (`hclwrite.Format`), since the
// Terraform specific rewrites don't apply to them. It returns nil when the
// text is unchanged.
func FormatText(filePath string, src []byte) ([]byte, error) {
	// like `tofu fmt`, the file has to parse before it's formatted, since the
	// formatter can make drastic changes to invalid code
	if _, diags := hclsyntax.ParseConfig(src, filePath, hcl.InitialPos); diags.HasErrors() {
		return nil, diagnosticsError(diags)
	}

	var output []byte
	if IsTerraformFile(filePath) {
		output = tofufmt.FormatSourceCode(src, filePath)
	} else {
		output = hclwrite.Format(src)
	}
	if bytes.Equal(output, src) {
		return nil, nil
	}
	return output, nil
}

// IsTerraformFile gets whether the file is one `tofu fmt` formats.
func IsTerraformFile(filePath string) bool {
	// dprint matches extensions case insensitively
	lowerPath := strings.ToLower(filePath)
	for _, ext := range tofufmt.SupportedExts {
		if strings.HasSuffix(lowerPath, ext) {
			return true
		}
	}
	return false
}

// diagnosticsError formats the error diagnostics as `file:line:col: summary;
// detail`, one per line.
func diagnosticsError(diags hcl.Diagnostics) error {
	var lines []string
	for _, diag := range diags {
		if diag.Severity != hcl.DiagError {
			continue
		}
		var line strings.Builder
		if diag.Subject != nil {
			fmt.Fprintf(&line, "%s:%d:%d: ", diag.Subject.Filename, diag.Subject.Start.Line, diag.Subject.Start.Column)
		}
		line.WriteString(diag.Summary)
		if diag.Detail != "" {
			line.WriteString("; ")
			line.WriteString(diag.Detail)
		}
		lines = append(lines, line.String())
	}
	return errors.New(strings.Join(lines, "\n"))
}
