import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const extensionDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(extensionDir, "..", "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(
  extensionDir,
  "syntaxes",
  "value-dsl.tmLanguage.json",
);

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));
const tokensByCategory = groupTokensByCategory(spec.tokens || []);
const enumTokens = collectEnumTokens(spec);

const grammar = {
  $schema:
    "https://raw.githubusercontent.com/martinring/tmlanguage/master/tmlanguage.json",
  name: "Value DSL",
  scopeName: "source.dsl",
  fileTypes: ["dsl"],
  patterns: [
    { include: "#comments" },
    { include: "#strings" },
    { include: "#numbers" },
    { include: "#operators" },
    { include: "#clauses" },
    { include: "#metadata" },
    { include: "#keywords" },
    { include: "#verbs" },
    { include: "#constants" },
    { include: "#identifiers" },
  ],
  repository: {
    comments: {
      patterns: [
        {
          name: "comment.line.double-slash.dsl",
          match: `${escapeRegex(spec.line_comment_prefix || "//")}.*$`,
        },
      ],
    },
    strings: {
      patterns: [
        {
          name: "string.quoted.dsl",
          match: spec.string_pattern || String.raw`(?:"[^"]*"|'[^']*')`,
        },
      ],
    },
    numbers: {
      patterns: [
        {
          name: "constant.numeric.dsl",
          match: boundaryPattern(spec.number_pattern),
        },
      ],
    },
    operators: punctuationPattern(
      "keyword.operator.dsl",
      spec.punctuation_tokens || [],
    ),
    clauses: tokenPattern(
      "keyword.control.clause.dsl",
      tokensByCategory.clause || [],
    ),
    metadata: tokenPattern(
      "keyword.other.metadata.dsl",
      tokensByCategory.metadata || [],
    ),
    keywords: tokenPattern(
      "keyword.other.dsl",
      tokensByCategory.keyword || [],
    ),
    verbs: tokenPattern(
      "support.function.builtin.dsl",
      tokensByCategory.verb || [],
    ),
    constants: tokenPattern("constant.language.dsl", enumTokens),
    identifiers: {
      patterns: [
        {
          name: "variable.other.identifier.dsl",
          match: boundaryPattern(spec.identifier_pattern),
        },
      ],
    },
  },
};

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, JSON.stringify(grammar, null, 2) + "\n");

function groupTokensByCategory(tokens) {
  const grouped = {};
  for (const token of tokens) {
    if (!token.category || !token.text) continue;
    grouped[token.category] ||= [];
    grouped[token.category].push(token.text);
  }
  for (const category of Object.keys(grouped)) {
    grouped[category] = [...new Set(grouped[category])].sort();
  }
  return grouped;
}

function collectEnumTokens(spec) {
  const out = new Set();
  for (const decl of spec.declarations || []) {
    collectEnumsFromMatchers(decl.header || [], out);
    for (const line of decl.body?.lines || []) {
      collectEnumsFromMatchers(line.pattern || [], out);
    }
  }
  return [...out].sort();
}

function collectEnumsFromMatchers(matchers, out) {
  for (const matcher of matchers || []) {
    switch (matcher.type) {
      case "enum":
        for (const value of matcher.values || []) {
          out.add(value);
        }
        break;
      case "list":
        collectEnumsFromMatchers([matcher.item], out);
        break;
      case "opt":
        collectEnumsFromMatchers(matcher.inner || [], out);
        break;
      default:
        break;
    }
  }
}

function tokenPattern(scope, values) {
  const uniqueValues = [...new Set(values || [])].filter(Boolean).sort();
  return {
    patterns:
      uniqueValues.length === 0
        ? []
        : [
            {
              name: scope,
              match: boundaryPattern(uniqueValues.map(escapeRegex).join("|")),
            },
          ],
  };
}

function punctuationPattern(scope, values) {
  const uniqueValues = [...new Set(values || [])].filter(Boolean).sort();
  return {
    patterns:
      uniqueValues.length === 0
        ? []
        : [
            {
              name: scope,
              match: uniqueValues.map(escapeRegex).join("|"),
            },
          ],
  };
}

function boundaryPattern(pattern) {
  return `(?<![A-Za-z0-9_-])(?:${pattern})(?![A-Za-z0-9_-])`;
}

function escapeRegex(value) {
  return String(value).replace(/[\\^$.*+?()[\]{}|]/g, "\\$&");
}
