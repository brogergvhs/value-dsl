import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(rootDir, "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(rootDir, "generated", "grammar-fragments.js");

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));

const content = `// AUTO-GENERATED from generated/grammar.json by scripts/generate-grammar-fragments.mjs
module.exports = ${JSON.stringify(
  {
    identifierPattern: spec.identifier_pattern,
    numberPattern: spec.number_pattern,
    stringPattern: spec.string_pattern,
    lineCommentPrefix: spec.line_comment_prefix || "//",
    reservedKeywords: spec.reserved_keywords || [],
    tokenKindRules: spec.token_kind_rules || [],
    actionSystem: spec.action_system || null,
    declarations: spec.declarations || [],
    actions: spec.actions || [],
  },
  null,
  2,
)};
`;

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, content);
