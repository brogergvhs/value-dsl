import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(rootDir, "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(rootDir, "queries", "highlights.scm");

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));

const punctuationTokens = spec.punctuation_tokens || [];
const punctuationSet = new Set(punctuationTokens);
const keywordTokens = collectKeywordTokens(spec).filter(
  (token) => !punctuationSet.has(token),
);

const lines = [
  "; AUTO-GENERATED from generated/grammar.json by scripts/generate-highlights.mjs",
  "",
  "(comment) @comment",
  "(string_literal) @string",
  renderTokenList(punctuationTokens, "@operator"),
  "",
  renderTokenList(keywordTokens, "@keyword"),
  ...collectStructuralCaptures(spec),
  "",
].filter(Boolean);

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, lines.join("\n"));

function collectKeywordTokens(spec) {
  const out = new Set();

  for (const decl of spec.declarations || []) {
    collectKeywordsFromMatchers(decl.header || [], out);
    if (decl.body?.lines) {
      for (const line of decl.body.lines) {
        collectKeywordsFromMatchers(line.pattern || [], out);
      }
    }
  }

  return [...out].sort();
}

function collectKeywordsFromMatchers(matchers, out) {
  for (const m of matchers || []) {
    switch (m.type) {
      case "kw":
        for (const token of String(m.text || "")
          .split(/\s+/)
          .filter(Boolean)) {
          out.add(token);
        }
        break;
      case "opt":
        collectKeywordsFromMatchers(m.inner || [], out);
        break;
      case "list":
        if (m.separator) {
          out.add(m.separator);
        }
        if (m.item) {
          collectKeywordsFromMatchers([m.item], out);
        }
        break;
      default:
        break;
    }
  }
}

function renderTokenList(values, capture) {
  const uniqueValues = [...new Set(values)].sort();
  if (uniqueValues.length === 0) {
    return "";
  }
  return `[${uniqueValues.map((value) => JSON.stringify(value)).join(" ")}] ${capture}`;
}

function collectStructuralCaptures(spec) {
  const actionSystem = spec.action_system;
  const tokenKindRules = Object.fromEntries(
    (spec.token_kind_rules || []).map((tkr) => [tkr.kind, tkr]),
  );
  const out = [];

  for (const decl of spec.declarations || []) {
    if (!decl.body?.lines) continue;
    for (const line of decl.body.lines) {
      if (actionSystem && line.node_name === actionSystem.clause_node_name) {
        continue;
      }
      walkPatternForCaptures(line.pattern || [], line.node_name, tokenKindRules, out);
    }
  }

  if (actionSystem) {
    for (const form of actionSystem.forms || []) {
      walkPatternForCaptures(form.pattern || [], form.node_name, tokenKindRules, out);
    }
  }

  return out;
}

function walkPatternForCaptures(matchers, nodeName, tokenKindRules, out) {
  for (const m of matchers || []) {
    switch (m.type) {
      case "tok": {
        if (!m.name) break;
        const tkr = tokenKindRules[m.kind];
        if (!tkr) break;
        if (tkr.values?.length > 0) {
          out.push(`(${nodeName} ${m.name}: (${tkr.rule}) @function.builtin)`);
        }
        if (tkr.fallback) {
          out.push(`(${nodeName} ${m.name}: (${tkr.fallback}) @function.builtin)`);
        } else if (!tkr.values?.length) {
          out.push(`(${nodeName} ${m.name}: (${tkr.rule}) @function.builtin)`);
        }
        break;
      }
      case "enum":
        if (m.name) {
          out.push(`(${nodeName} ${m.name}: (identifier) @constant)`);
        }
        break;
      case "opt":
        walkPatternForCaptures(m.inner || [], nodeName, tokenKindRules, out);
        break;
      case "list":
        if (m.item) {
          walkPatternForCaptures([m.item], nodeName, tokenKindRules, out);
        }
        break;
      default:
        break;
    }
  }
}
