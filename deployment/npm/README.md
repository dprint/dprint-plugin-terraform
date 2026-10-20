# @dprint/terraform

npm distribution of [dprint-plugin-terraform](https://github.com/dprint/dprint-plugin-terraform), which formats Terraform and OpenTofu files like `tofu fmt` does, and other HCL files like `hclfmt` does.

Use this with [@dprint/formatter](https://github.com/dprint/js-formatter) or just use @dprint/formatter and download the [dprint-plugin-terraform Wasm file](https://github.com/dprint/dprint-plugin-terraform/releases).

## Example

```ts
import { createFromBuffer } from "@dprint/formatter";
import { getBuffer } from "@dprint/terraform";

const formatter = createFromBuffer(getBuffer());

console.log(formatter.formatText({
  filePath: "main.tf",
  fileText: "variable \"name\" {\ntype = \"string\"\n}\n",
}));
```
