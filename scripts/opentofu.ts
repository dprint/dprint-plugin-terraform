/**
 * Checks that the `tofu fmt` logic ported to `internal/tofufmt/fmt.go` is
 * still the same in the latest OpenTofu release. The port is updated by hand,
 * so this only reports when it has to be.
 *
 * The fingerprint covers the formatting functions of OpenTofu's
 * `internal/command/fmt.go`, the extensions `tofu fmt` formats, and the HCL
 * version in OpenTofu's go.mod (OpenTofu uses its own fork of HCL), since a
 * change to any of them can change the output of `tofu fmt`.
 *
 * Run it directly to print the fingerprint of the latest release:
 *
 * ```sh
 * deno run -A ./scripts/opentofu.ts
 * ```
 */
import { $ } from "automation";

/** The fingerprint of the OpenTofu version in `internal/tofufmt/OPENTOFU_VERSION`. */
const PORTED_FINGERPRINT = "a758e5a2c57e1eff4e602a7e9d2da3973ea0916ca559d982e75b52a6686e684f";

const rootDirPath = $.path(import.meta.dirname!).parentOrThrow();
const versionFilePath = rootDirPath.join("internal/tofufmt/OPENTOFU_VERSION");

const portedFunctions = [
  "formatSourceCode",
  "formatBody",
  "formatValueExpr",
  "formatTypeExpr",
  "trimNewlines",
];

if (import.meta.main) {
  const version = await getLatestOpenTofuVersion();
  $.log(`OpenTofu ${version}: ${await getFingerprint(await getFormattingSource(version))}`);
}

/**
 * Gets an error message when the formatting logic of the latest OpenTofu
 * release differs from the ported version, with a diff of the change.
 */
export async function checkOpenTofuFmt(): Promise<string | undefined> {
  const portedVersion = getPortedOpenTofuVersion();
  const latestVersion = await getLatestOpenTofuVersion();
  $.logLight(`Ported OpenTofu version: ${portedVersion}, latest: ${latestVersion}`);
  const latestSource = await getFormattingSource(latestVersion);
  if (await getFingerprint(latestSource) === PORTED_FINGERPRINT) {
    return undefined;
  }

  const diff = await getDiff(await getFormattingSource(portedVersion), latestSource);
  return [
    `The formatting logic of \`tofu fmt\` changed between OpenTofu ${portedVersion} and ${latestVersion}.`,
    `Port the change to internal/tofufmt/fmt.go by hand, then update internal/tofufmt/OPENTOFU_VERSION`,
    `and PORTED_FINGERPRINT in scripts/opentofu.ts (\`deno run -A ./scripts/opentofu.ts\` prints it).`,
    ``,
    diff,
  ].join("\n");
}

/** Gets the OpenTofu version `internal/tofufmt/fmt.go` was ported from. */
export function getPortedOpenTofuVersion(): string {
  return versionFilePath.readTextSync().trim();
}

async function getLatestOpenTofuVersion() {
  const release = await $.request("https://api.github.com/repos/opentofu/opentofu/releases/latest")
    .header(getGitHubHeaders())
    .json<{ tag_name?: string }>();
  const version = release.tag_name?.replace(/^v/, "");
  if (version == null) {
    throw new Error("Could not find the latest OpenTofu release.");
  }
  return version;
}

/**
 * Gets the parts of OpenTofu's source that decide the output of `tofu fmt`.
 * It throws when one is missing, since then the source moved and has to be
 * looked at by hand.
 */
async function getFormattingSource(version: string) {
  const baseUrl = `https://raw.githubusercontent.com/opentofu/opentofu/v${version}`;
  const fmtSource = await $.request(`${baseUrl}/internal/command/fmt.go`).text();
  const goMod = await $.request(`${baseUrl}/go.mod`).text();

  const parts = [extractBlock(fmtSource, /^\tfmtSupportedExts = \[\]string\{$/m, "\t}", "fmtSupportedExts")];
  for (const name of portedFunctions) {
    const pattern = new RegExp(`^func (\\(c \\*FmtCommand\\) )?${name}\\(.*\\{$`, "m");
    parts.push(extractBlock(fmtSource, pattern, "}", name));
  }
  const hclLines = goMod.split("\n").filter((line) => line.includes("github.com/hashicorp/hcl/v2"));
  if (hclLines.length === 0) {
    throw new Error(`Could not find the HCL version in the go.mod of OpenTofu ${version}.`);
  }
  parts.push(hclLines.map((line) => line.trim()).join("\n"));
  return parts.join("\n\n") + "\n";
}

/** Extracts the text from the line matching the pattern to the closing line. */
function extractBlock(source: string, startPattern: RegExp, closingLine: string, name: string) {
  const lines = source.replaceAll("\r\n", "\n").split("\n");
  const start = lines.findIndex((line) => startPattern.test(line));
  const end = lines.findIndex((line, i) => i > start && line === closingLine);
  if (start === -1 || end === -1) {
    throw new Error(`Could not find \`${name}\` in OpenTofu's internal/command/fmt.go.`);
  }
  return lines.slice(start, end + 1).join("\n");
}

async function getFingerprint(text: string) {
  const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text));
  return Array.from(new Uint8Array(hash), (byte) => byte.toString(16).padStart(2, "0")).join("");
}

async function getDiff(oldText: string, newText: string) {
  const tempDir = $.path(await Deno.makeTempDir());
  try {
    tempDir.join("ported").writeTextSync(oldText);
    tempDir.join("latest").writeTextSync(newText);
    // exits with 1 when there's a difference
    return await $`git diff --no-index --no-color ported latest`.cwd(tempDir).noThrow().text();
  } finally {
    tempDir.removeSync({ recursive: true });
  }
}

function getGitHubHeaders(): Record<string, string> {
  const token = Deno.env.get("GITHUB_TOKEN");
  return token ? { authorization: `Bearer ${token}` } : {};
}
