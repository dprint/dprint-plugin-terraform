// @ts-check
const assert = require("assert");
const createFromBuffer = require("@dprint/formatter").createFromBuffer;
const getBuffer = require("./index").getBuffer;

const formatter = createFromBuffer(getBuffer());

assert.deepStrictEqual(formatter.getFileMatchingInfo().fileExtensions, ["tf", "tofu", "tfvars", "hcl"]);

// terraform files get the `tofu fmt` rewrites
assert.strictEqual(
  formatter.formatText({
    filePath: "main.tf",
    fileText: "variable \"a\" {\ntype = list\ndefault = \"${local.a}\"\n}\n",
  }),
  "variable \"a\" {\n  type    = list(any)\n  default = local.a\n}\n",
);

// as do terraform test files, which come in through the `hcl` extension
assert.strictEqual(
  formatter.formatText({
    filePath: "tests/main.tftest.hcl",
    fileText: "run \"a\" {\ncommand = plan\nvariables {\nname = \"${var.x}\"\n}\n}\n",
  }),
  "run \"a\" {\n  command = plan\n  variables {\n    name = var.x\n  }\n}\n",
);

// other HCL files are only formatted like `hclfmt` does
assert.strictEqual(
  formatter.formatText({
    filePath: "terragrunt.hcl",
    fileText: "inputs = {\nname = \"${local.name}\"\n}\n",
  }),
  "inputs = {\n  name = \"${local.name}\"\n}\n",
);

// the plugin has no options
formatter.setConfig({}, { indentWidth: 4 });
assert.deepStrictEqual(formatter.getConfigDiagnostics(), [{
  propertyName: "indentWidth",
  message: "Unknown property in configuration",
}]);
formatter.setConfig({}, {});

assert.throws(
  () => formatter.formatText({ filePath: "main.tf", fileText: "a = \n" }),
  /main\.tf:1:5: Invalid expression; Expected the start of an expression/,
);
