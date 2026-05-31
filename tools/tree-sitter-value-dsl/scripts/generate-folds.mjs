import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(rootDir, "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(rootDir, "queries", "folds.scm");

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));

const foldableNodes = (spec.declarations || [])
  .filter((decl) => decl.body)
  .map((decl) => decl.node_name);

const lines = [
  "; AUTO-GENERATED from generated/grammar.json by scripts/generate-folds.mjs",
  "",
  ...foldableNodes.map((nodeName) => `(${nodeName}) @fold`),
  "",
];

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, lines.join("\n"));
