package plugin

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFormatFixtures formats each `testdata/<case>/input.<ext>` file and
// compares it with `output.<ext>`, which is what `tofu fmt` (or `hclfmt` for
// generic HCL files) outputs for it. A case without an output file must be
// unchanged, and one with an `error.txt` must fail with that error.
func TestFormatFixtures(t *testing.T) {
	for _, inputPath := range fixtureInputs(t) {
		caseDir := filepath.Dir(inputPath)
		t.Run(filepath.Base(caseDir), func(t *testing.T) {
			input := readFile(t, inputPath)
			ext := strings.TrimPrefix(filepath.Base(inputPath), "input")
			output, err := FormatText(filepath.ToSlash(inputPath), input)

			if expectedError, ok := tryReadFile(t, filepath.Join(caseDir, "error.txt")); ok {
				if err == nil {
					t.Fatalf("expected an error, got output %q", output)
				}
				if err.Error() != strings.TrimSpace(string(expectedError)) {
					t.Fatalf("got error:\n%s\nwant:\n%s", err, expectedError)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expected, ok := tryReadFile(t, filepath.Join(caseDir, "output"+ext))
			if !ok {
				if output != nil {
					t.Fatalf("expected no change, got:\n%s", output)
				}
				return
			}
			if output == nil {
				t.Fatalf("expected a change to:\n%s", expected)
			}
			if !bytes.Equal(output, expected) {
				t.Fatalf("got:\n%s\nwant:\n%s", output, expected)
			}

			// formatting again shouldn't change anything
			again, err := FormatText(filepath.ToSlash(inputPath), output)
			if err != nil {
				t.Fatalf("unexpected error when formatting twice: %v", err)
			}
			if again != nil {
				t.Fatalf("formatting twice changed the text:\n%s", again)
			}
		})
	}
}

// TestMatchesTofu compares the output with the `tofu` (or `terraform`) binary
// in the TOFU environment variable, for the Terraform fixtures and the files
// in the TOFU_CORPUS directory when that's set.
func TestMatchesTofu(t *testing.T) {
	tofuPath := os.Getenv("TOFU")
	if tofuPath == "" {
		t.Skip("set TOFU to the path of a tofu binary to compare with it")
	}
	paths := fixtureInputs(t)
	if corpus := os.Getenv("TOFU_CORPUS"); corpus != "" {
		_ = filepath.WalkDir(corpus, func(path string, entry os.DirEntry, err error) error {
			if err == nil && !entry.IsDir() {
				paths = append(paths, path)
			}
			return nil
		})
	}

	compared := 0
	for _, path := range paths {
		if !IsTerraformFile(path) {
			continue
		}
		input := readFile(t, path)
		cmd := exec.Command(tofuPath, "fmt", "-no-color", "-")
		cmd.Stdin = bytes.NewReader(input)
		expected, tofuErr := cmd.Output()
		output, err := FormatText(path, input)
		if (tofuErr != nil) != (err != nil) {
			t.Errorf("%s: tofu error %v, plugin error %v", path, tofuErr, err)
			continue
		}
		if err != nil {
			continue
		}
		if output == nil {
			output = input
		}
		if !bytes.Equal(output, expected) {
			t.Errorf("%s: differs from tofu fmt\ngot:\n%s\nwant:\n%s", path, output, expected)
		}
		compared++
	}
	t.Logf("compared %d files with %s", compared, tofuPath)
}

func TestFormatPreservesCRLF(t *testing.T) {
	input := "variable \"a\" {\r\n  type = list\r\n  default = \"${local.x}\"\r\n}\r\n"
	output, err := FormatText("/repo/main.tf", []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	expected := "variable \"a\" {\r\n  type    = list(any)\r\n  default = local.x\r\n}\r\n"
	if string(output) != expected {
		t.Fatalf("got %q, want %q", output, expected)
	}
}

func TestFormatErrors(t *testing.T) {
	_, err := FormatText(`C:\repo\main.tf`, []byte("a = 1 2\n"))
	expected := `C:\repo\main.tf:1:7: Missing newline after argument; An argument definition must end with a newline.`
	if err == nil || err.Error() != expected {
		t.Fatalf("got error:\n%v\nwant:\n%s", err, expected)
	}

	// generic HCL files have to parse too
	_, err = FormatText("/repo/config.hcl", []byte("a = {\n"))
	expected = "/repo/config.hcl:2:1: Missing expression; Expected the start of an expression, but found the end of the file."
	if err == nil || err.Error() != expected {
		t.Fatalf("got error:\n%v\nwant:\n%s", err, expected)
	}
}

func TestIsTerraformFile(t *testing.T) {
	tests := map[string]bool{
		"/repo/main.tf":                    true,
		`C:\repo\main.tf`:                  true,
		"/repo/MAIN.TF":                    true,
		"/repo/main.tofu":                  true,
		"/repo/dev.tfvars":                 true,
		"/repo/tests/main.tftest.hcl":      true,
		"/repo/tests/main.tofutest.hcl":    true,
		"/repo/tests/main.TFTEST.HCL":      true,
		"/repo/terragrunt.hcl":             false,
		"/repo/.terraform.lock.hcl":        false,
		"/repo/tests/mocks/aws.tfmock.hcl": false,
		"/repo/build.pkr.hcl":              false,
		"/repo/main.tf.json":               false,
	}
	for path, expected := range tests {
		if got := IsTerraformFile(path); got != expected {
			t.Errorf("IsTerraformFile(%q) = %v, want %v", path, got, expected)
		}
	}
}

func fixtureInputs(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("testdata", "*", "input.*"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no fixtures found: %v", err)
	}
	return paths
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func tryReadFile(t *testing.T, path string) ([]byte, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatal(err)
	}
	return data, true
}
