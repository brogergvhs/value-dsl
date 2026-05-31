import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(rootDir, "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(rootDir, "queries", "commentstring.scm");

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));
const commentPrefix = spec.line_comment_prefix || "//";
const lines = [
  "; AUTO-GENERATED from generated/grammar.json by scripts/generate-commentstring.mjs",
  "",
  `((source_file) @comment`,
  ` (#set! commentstring "${commentPrefix} %s"))`,
  "",
];

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, lines.join("\n"));
