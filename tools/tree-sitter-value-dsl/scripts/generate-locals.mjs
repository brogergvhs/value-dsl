import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(rootDir, "..", "..");
const specPath = path.join(repoRoot, "generated", "grammar.json");
const outputPath = path.join(rootDir, "queries", "locals.scm");

const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));
const actionSystem = spec.action_system;

const definitionPatterns = [];
const referencePatterns = [];

for (const decl of spec.declarations || []) {
  const declNode = decl.node_name;

  const headerDefinitions = collectTopLevelDefinitions(decl);
  for (const fieldName of headerDefinitions) {
    definitionPatterns.push(
      `(${declNode} ${fieldName}: (identifier) @local.definition)`,
    );
  }

  const headerReferences = collectReferences(decl.header || []);
  for (const fieldName of headerReferences) {
    referencePatterns.push(
      `(${declNode} ${fieldName}: (identifier) @local.reference)`,
    );
  }

  for (const line of decl.body?.lines || []) {
    if (actionSystem && line.node_name === actionSystem.clause_node_name) {
      for (const form of actionSystem.forms || []) {
        const formRefs = collectReferences(form.pattern || []);
        for (const fieldName of formRefs) {
          referencePatterns.push(
            `(${actionSystem.clause_node_name} (${actionSystem.expr_node_name} (${form.node_name} ${fieldName}: (identifier) @local.reference)))`,
          );
        }
      }
      continue;
    }

    const lineRefs = collectReferences(line.pattern || []);
    for (const fieldName of lineRefs) {
      referencePatterns.push(
        `(${line.node_name} ${fieldName}: (identifier) @local.reference)`,
      );
    }
  }
}

const lines = [
  "; AUTO-GENERATED from generated/grammar.json by scripts/generate-locals.mjs",
  "",
  "(source_file) @local.scope",
  "",
  "; Definitions",
  ...unique(definitionPatterns),
  "",
  "; References",
  ...unique(referencePatterns),
  "",
];

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, lines.join("\n"));

function collectTopLevelDefinitions(decl) {
  const out = [];
  const header = decl.header || [];

  for (let i = 0; i < header.length; i++) {
    const matcher = header[i];
    const hasLeadingKeyword = header.slice(0, i).some((m) => m.type === "kw");
    if (!hasLeadingKeyword) continue;

    if (matcher.type === "ident" && matcher.name) {
      out.push(matcher.name);
      break;
    }
    if (matcher.type === "list" && matcher.name) {
      out.push(matcher.name);
      break;
    }
  }

  return out;
}

function collectReferences(matchers) {
  const out = [];
  collectReferencesInto(matchers, out);
  return out;
}

function collectReferencesInto(matchers, out) {
  for (const m of matchers || []) {
    switch (m.type) {
      case "ref":
        if (m.name) {
          out.push(m.name);
        }
        break;
      case "list":
        if (m.item) {
          collectReferencesInto([m.item], out);
        }
        break;
      case "opt":
        collectReferencesInto(m.inner || [], out);
        break;
      default:
        break;
    }
  }
}

function unique(values) {
  return [...new Set(values)].sort();
}
