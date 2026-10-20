# dprint-plugin-terraform

[![npm version](https://img.shields.io/npm/v/@dprint/terraform.svg)](https://www.npmjs.com/package/@dprint/terraform) [![CI](https://github.com/dprint/dprint-plugin-terraform/workflows/CI/badge.svg)](https://github.com/dprint/dprint-plugin-terraform/actions?query=workflow%3ACI)

Terraform, OpenTofu and HCL formatting plugin for [dprint](https://github.com/dprint/dprint).

It formats Terraform and OpenTofu files byte for byte like `tofu fmt` does, and other HCL files, like Packer, Nomad and Terragrunt configuration, like `hclfmt` does. It's [HCL](https://github.com/hashicorp/hcl)'s own Go code compiled to Wasm with [TinyGo](https://tinygo.org), along with the formatting logic of `tofu fmt` ported from [OpenTofu](https://github.com/opentofu/opentofu).

## Install

[Install](https://dprint.dev/install/) and [setup](https://dprint.dev/setup/) dprint.

Then in your project's directory with a dprint.json file, run:

```shellsession
dprint add terraform
```

Note: You do not need Terraform or OpenTofu installed, as dprint runs the plugin from the .wasm file in a sandboxed environment.

## Files

| Files                                                                                      | Formatted like                                                                                                                                                                             |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `.tf`, `.tofu`, `.tfvars`, `.tftest.hcl` and `.tofutest.hcl`                               | `tofu fmt`, which on top of HCL's formatting unwraps redundant interpolations (`"${var.a}"` to `var.a`), quotes block labels, and normalizes legacy variable types (`list` to `list(any)`) |
| Other `.hcl` files (ex. Packer, Nomad and Terragrunt configuration, `.terraform.lock.hcl`) | `hclfmt`, which is HCL's formatting only, since the Terraform specific rewrites don't apply to them                                                                                        |

dprint matches files by their last extension, so the plugin formats `.hcl` files by default to also format `.tftest.hcl` and `.tofutest.hcl` files. To leave some `.hcl` files alone, add them to `excludes` in your dprint config, or use the plugin's [`associations`](https://dprint.dev/config/#associations).

A file has to parse before it's formatted, like `tofu fmt` requires, and a file that doesn't is reported with HCL's error (ex. `main.tf:2:8: Invalid expression; Expected the start of an expression, but found an invalid expression token.`). JSON files (`.tf.json` and `.tfvars.json`) aren't formatted, the same as with `tofu fmt`.

## Configuration

There's nothing to configure, since `tofu fmt` and `terraform fmt` have no style options. The plugin keeps the line endings of the file the same way they do, so it doesn't use dprint's `newLineKind`.

## JS Formatting API

- [JS Formatter](https://github.com/dprint/js-formatter) - Browser/Deno and Node
- [npm package](https://www.npmjs.com/package/@dprint/terraform)

## Versioning

This repo automatically upgrades to the latest version of HCL once a day. You can check which version of HCL is being used by looking at the `github.com/hashicorp/hcl/v2` entry in the go.mod file in this repo:

https://github.com/dprint/dprint-plugin-terraform/blob/main/go.mod

The `tofu fmt` logic in `internal/tofufmt/fmt.go` is ported by hand, from the OpenTofu version in [`internal/tofufmt/OPENTOFU_VERSION`](internal/tofufmt/OPENTOFU_VERSION). The daily check fails when the formatting logic of the latest OpenTofu release changes, so the change can be ported.

At the moment, the version of this plugin does not reflect the version of HCL or OpenTofu.

## Development

The plugin is built with [TinyGo](https://tinygo.org), which produces a much smaller module than the Go compiler, and one that has no imports, so it doesn't rely on the dprint CLI providing WASI.

```shellsession
go test ./...
deno run -A ./scripts/build.ts # outputs target/plugin.wasm
```

The build uses a local `tinygo` (which needs binaryen's `wasm-opt`) when there is one, and otherwise the TinyGo Docker image.

The expected outputs in `internal/plugin/testdata` are what `tofu fmt` (and `hclfmt` for the generic `.hcl` fixture) output. To also compare the plugin with `tofu fmt` on the fixtures, and optionally a directory of your own files, set `TOFU`:

```shellsession
TOFU=tofu TOFU_CORPUS=path/to/terraform go test ./internal/plugin -run TestMatchesTofu -v
```

## Credits

- [HCL](https://github.com/hashicorp/hcl) by HashiCorp, under the Mozilla Public License 2.0.
- The formatting logic of `tofu fmt` is ported from [OpenTofu](https://github.com/opentofu/opentofu), under the Mozilla Public License 2.0. No code comes from Terraform, which is under the Business Source License since 1.6.
- The plugin interface is adapted from [dprint-plugin-shfmt](https://github.com/dprint/dprint-plugin-shfmt), which adapted it from Jake Bailey's [dprint-plugin-gofumpt](https://github.com/jakebailey/dprint-plugin-gofumpt).

## License

The plugin is licensed under the MIT License, except for `internal/tofufmt/fmt.go`, which is ported from OpenTofu and is licensed under the Mozilla Public License 2.0, like the file it's ported from. The compiled plugin also includes HCL, which is licensed under the Mozilla Public License 2.0, and its source code is available at https://github.com/hashicorp/hcl. See the [LICENSE](LICENSE) file for details, and [third_party](third_party) for the licenses of the code the plugin includes.
