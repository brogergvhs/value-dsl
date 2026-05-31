const fragments = require("./generated/grammar-fragments");

function keywordChoice(values) {
  return choice(...values);
}

// ---- Token kind dispatch ----

const tokenKindRuleMap = Object.fromEntries(
  (fragments.tokenKindRules || []).map((tkr) => [tkr.kind, tkr]),
);

function bareMatcherToRule($, matcher) {
  switch (matcher.type) {
    case "kw": {
      const parts = String(matcher.text || "")
        .split(/\s+/)
        .filter(Boolean);
      return parts.length === 1 ? parts[0] : seq(...parts);
    }

    case "ident":
    case "ref":
      return $.identifier;

    case "tok": {
      const tkr = tokenKindRuleMap[matcher.kind];
      if (!tkr) {
        return keywordChoice(matcher.values || []);
      }
      const primary = $[tkr.rule];
      return tkr.fallback ? choice(primary, $[tkr.fallback]) : primary;
    }

    case "str":
      return choice($.identifier, $.string_literal);

    case "enum":
      return alias(keywordChoice(matcher.values || []), $.identifier);

    case "list": {
      const itemRule = bareMatcherToRule($, matcher.item);
      const fieldName = matcher.name || matcher.item?.name;
      const item = fieldName ? field(fieldName, itemRule) : itemRule;
      const repeated = fieldName ? field(fieldName, itemRule) : itemRule;
      return seq(item, repeat(seq(matcher.separator, repeated)));
    }

    case "opt":
      return optional(
        seq(...(matcher.inner || []).map((m) => matcherToRule($, m))),
      );

    case "free":
      return repeat1(choice($.identifier, $.number, $.string_literal));

    default:
      throw new Error(`unsupported matcher type: ${matcher.type}`);
  }
}

function matcherToRule($, matcher) {
  const rule = bareMatcherToRule($, matcher);

  if (matcher.type === "kw") {
    return rule;
  }

  if (matcher.type === "list" || matcher.type === "opt") {
    return rule;
  }

  if (matcher.name) {
    return field(matcher.name, rule);
  }

  return rule;
}

function buildPatternRule($, pattern) {
  return seq(...(pattern || []).map((m) => matcherToRule($, m)));
}

// ---- Declaration & line rules ----

function buildLineOccurrence($, line) {
  const nodeName = line.node_name;
  const node = field(nodeName, $[nodeName]);

  if (line.repeatable && line.required) {
    return repeat1(node);
  }
  if (line.repeatable) {
    return repeat(node);
  }
  if (line.required) {
    return node;
  }
  return optional(node);
}

function declarationBodyLines(decl) {
  const byNodeName = new Map();

  for (const line of decl.body?.lines || []) {
    const nodeName = line.node_name;
    const existing = byNodeName.get(nodeName);
    if (existing) {
      existing.required = existing.required || !!line.required;
      existing.repeatable = existing.repeatable || !!line.repeatable;
      continue;
    }
    byNodeName.set(nodeName, {
      ...line,
      required: !!line.required,
      repeatable: !!line.repeatable,
    });
  }

  return [...byNodeName.values()];
}

function declarationRule($, decl) {
  const header = buildPatternRule($, decl.header || []);

  if (!decl.body) {
    return header;
  }

  const body = declarationBodyLines(decl).map((line) =>
    buildLineOccurrence($, line),
  );
  return seq(header, ...body);
}

function generatedDeclarationRuleEntries() {
  return Object.fromEntries(
    (fragments.declarations || []).map((decl) => [
      decl.node_name,
      ($) => declarationRule($, decl),
    ]),
  );
}

function generatedLineRuleEntries() {
  const entries = [];
  const actionSystem = fragments.actionSystem;

  for (const decl of fragments.declarations || []) {
    for (const line of decl.body?.lines || []) {
      const nodeName = line.node_name;

      if (actionSystem && nodeName === actionSystem.clause_node_name) {
        entries.push([
          nodeName,
          ($) =>
            seq(
              ...buildPrefixRule($, actionSystem.prefix),
              field("action", $[actionSystem.expr_node_name]),
            ),
        ]);
        continue;
      }

      entries.push([nodeName, ($) => buildPatternRule($, line.pattern || [])]);
    }
  }

  return Object.fromEntries(entries);
}

function buildPrefixRule($, prefix) {
  return (prefix || []).map((m) => matcherToRule($, m));
}

// ---- Action variant rules ----

function generatedActionVariantRules() {
  const actionSystem = fragments.actionSystem;
  if (!actionSystem || !actionSystem.forms?.length) {
    return {};
  }

  const out = {};

  for (const form of actionSystem.forms) {
    out[form.node_name] = ($) => buildPatternRule($, form.pattern);
  }

  out[actionSystem.expr_node_name] = ($) => {
    const variants = actionSystem.forms.map((form) => $[form.node_name]);
    return choice(...variants);
  };

  return out;
}

// ---- Token kind dedicated rules ----

function generatedTokenKindRuleEntries() {
  const out = {};
  for (const tkr of fragments.tokenKindRules || []) {
    if (tkr.values?.length > 0) {
      out[tkr.rule] = () => keywordChoice(tkr.values);
    }
  }
  return out;
}

// ---- Assemble grammar ----

const generatedDeclarationRules = generatedDeclarationRuleEntries();
const generatedLineRules = generatedLineRuleEntries();
const generatedActionVariantRuleMap = generatedActionVariantRules();
const generatedTokenKindRules = generatedTokenKindRuleEntries();
const topLevelDeclarationNames = (fragments.declarations || []).map(
  (d) => d.node_name,
);

module.exports = grammar({
  name: "value_dsl",

  extras: ($) => [/\s+/, $.comment],

  word: ($) => $.identifier,

  rules: {
    source_file: ($) => repeat($._declaration),

    comment: () => token(seq(fragments.lineCommentPrefix || "//", /.*/)),

    identifier: () => new RegExp(fragments.identifierPattern),

    number: () => new RegExp(fragments.numberPattern),

    string_literal: () => new RegExp(fragments.stringPattern),

    _declaration: ($) =>
      choice(...topLevelDeclarationNames.map((name) => $[name])),

    ...generatedTokenKindRules,
    ...generatedDeclarationRules,
    ...generatedLineRules,
    ...generatedActionVariantRuleMap,
  },
});
