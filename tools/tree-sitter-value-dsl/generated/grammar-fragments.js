// AUTO-GENERATED from generated/grammar.json by scripts/generate-grammar-fragments.mjs
module.exports = {
  "identifierPattern": "[a-zA-Z_][a-zA-Z0-9_-]*",
  "numberPattern": "-?(?:[0-9]+(?:\\.[0-9]+)?|\\.[0-9]+)",
  "stringPattern": "(?:\"[^\"]*\"|'[^']*')",
  "lineCommentPrefix": "//",
  "reservedKeywords": [
    "access",
    "alert",
    "anonymize",
    "archive",
    "assignment",
    "capture",
    "collect",
    "crosses",
    "delete",
    "detect",
    "enters",
    "exceeds",
    "if",
    "inform",
    "leaves",
    "limit",
    "linked_to",
    "log",
    "mask",
    "measure",
    "monitor",
    "notify",
    "observe",
    "of",
    "priority",
    "record",
    "remove",
    "reports",
    "requests",
    "requirement",
    "restrict",
    "retain",
    "retention",
    "save",
    "shall",
    "shows",
    "stakeholder",
    "stakeholders",
    "starts",
    "stops",
    "store",
    "system",
    "track",
    "using",
    "value",
    "warn",
    "when",
    "where",
    "while"
  ],
  "tokenKindRules": [
    {
      "kind": "ident",
      "rule": "identifier"
    },
    {
      "kind": "number",
      "rule": "number"
    },
    {
      "kind": "string",
      "rule": "string_literal"
    },
    {
      "kind": "when_verb",
      "rule": "when_verb",
      "fallback": "identifier",
      "values": [
        "enters",
        "leaves",
        "starts",
        "stops",
        "requests",
        "crosses",
        "exceeds",
        "shows",
        "reports"
      ]
    },
    {
      "kind": "action_verb",
      "rule": "identifier"
    }
  ],
  "actionSystem": {
    "clause_node_name": "action_clause",
    "expr_node_name": "action_expression",
    "prefix": [
      {
        "type": "kw",
        "text": "system"
      },
      {
        "type": "kw",
        "text": "shall"
      }
    ],
    "forms": [
      {
        "form": "object_of_target",
        "node_name": "targeted_action",
        "pattern": [
          {
            "type": "tok",
            "name": "verb",
            "kind": "action_verb",
            "values": [
              "track",
              "monitor",
              "observe",
              "detect",
              "measure",
              "record",
              "collect",
              "capture"
            ]
          },
          {
            "type": "str",
            "name": "object"
          },
          {
            "type": "kw",
            "text": "of"
          },
          {
            "type": "ref",
            "name": "target",
            "kind": "stakeholder"
          },
          {
            "type": "opt",
            "inner": [
              {
                "type": "kw",
                "text": "using"
              },
              {
                "type": "str",
                "name": "mechanism"
              }
            ]
          }
        ]
      },
      {
        "form": "direct_target",
        "node_name": "direct_action",
        "pattern": [
          {
            "type": "tok",
            "name": "verb",
            "kind": "action_verb",
            "values": [
              "notify",
              "alert",
              "inform",
              "warn"
            ]
          },
          {
            "type": "ref",
            "name": "target",
            "kind": "stakeholder"
          },
          {
            "type": "opt",
            "inner": [
              {
                "type": "kw",
                "text": "using"
              },
              {
                "type": "str",
                "name": "mechanism"
              }
            ]
          }
        ]
      }
    ]
  },
  "declarations": [
    {
      "kind": "stakeholder",
      "node_name": "stakeholder_decl",
      "keyword": "stakeholder",
      "header": [
        {
          "type": "kw",
          "text": "stakeholder"
        },
        {
          "type": "list",
          "name": "names",
          "separator": ",",
          "item": {
            "type": "ident",
            "name": "name"
          }
        }
      ]
    },
    {
      "kind": "value",
      "node_name": "value_decl",
      "keyword": "value",
      "header": [
        {
          "type": "kw",
          "text": "value"
        },
        {
          "type": "ident",
          "name": "name"
        },
        {
          "type": "opt",
          "inner": [
            {
              "type": "kw",
              "text": "="
            },
            {
              "type": "tok",
              "name": "angle",
              "kind": "number"
            },
            {
              "type": "kw",
              "text": ","
            },
            {
              "type": "tok",
              "name": "radius",
              "kind": "number"
            }
          ]
        }
      ]
    },
    {
      "kind": "requirement",
      "node_name": "requirement_decl",
      "keyword": "requirement",
      "header": [
        {
          "type": "kw",
          "text": "requirement"
        },
        {
          "type": "ident",
          "name": "id"
        }
      ],
      "body": {
        "lines": [
          {
            "name": "While",
            "node_name": "while_clause",
            "clause_kind": "while",
            "required": false,
            "repeatable": true,
            "order": 1,
            "pattern": [
              {
                "type": "kw",
                "text": "while"
              },
              {
                "type": "ref",
                "name": "actor",
                "kind": "stakeholder"
              },
              {
                "type": "free",
                "name": "rest"
              }
            ]
          },
          {
            "name": "When",
            "node_name": "when_clause",
            "clause_kind": "when",
            "required": false,
            "repeatable": false,
            "order": 2,
            "pattern": [
              {
                "type": "kw",
                "text": "when"
              },
              {
                "type": "ref",
                "name": "actor",
                "kind": "stakeholder"
              },
              {
                "type": "tok",
                "name": "verb",
                "kind": "when_verb",
                "values": [
                  "enters",
                  "leaves",
                  "starts",
                  "stops",
                  "requests",
                  "crosses",
                  "exceeds",
                  "shows",
                  "reports"
                ]
              },
              {
                "type": "ident",
                "name": "target"
              }
            ]
          },
          {
            "name": "If",
            "node_name": "if_clause",
            "clause_kind": "if",
            "required": false,
            "repeatable": false,
            "order": 3,
            "pattern": [
              {
                "type": "kw",
                "text": "if"
              },
              {
                "type": "ref",
                "name": "actor",
                "kind": "stakeholder"
              },
              {
                "type": "free",
                "name": "rest"
              }
            ]
          },
          {
            "name": "Where",
            "node_name": "where_clause",
            "clause_kind": "where",
            "required": false,
            "repeatable": false,
            "order": 4,
            "pattern": [
              {
                "type": "kw",
                "text": "where"
              },
              {
                "type": "free",
                "name": "rest"
              }
            ]
          },
          {
            "name": "Shall_monitoring",
            "node_name": "action_clause",
            "clause_kind": "system_shall",
            "required": true,
            "repeatable": false,
            "order": 5,
            "pattern": [
              {
                "type": "kw",
                "text": "system"
              },
              {
                "type": "kw",
                "text": "shall"
              },
              {
                "type": "tok",
                "name": "verb",
                "kind": "action_verb",
                "values": [
                  "track",
                  "monitor",
                  "observe",
                  "detect",
                  "measure",
                  "record",
                  "collect",
                  "capture"
                ]
              },
              {
                "type": "str",
                "name": "object"
              },
              {
                "type": "kw",
                "text": "of"
              },
              {
                "type": "ref",
                "name": "target",
                "kind": "stakeholder"
              },
              {
                "type": "opt",
                "inner": [
                  {
                    "type": "kw",
                    "text": "using"
                  },
                  {
                    "type": "str",
                    "name": "mechanism"
                  }
                ]
              }
            ]
          },
          {
            "name": "Shall_notification",
            "node_name": "action_clause",
            "clause_kind": "system_shall",
            "required": true,
            "repeatable": false,
            "order": 5,
            "pattern": [
              {
                "type": "kw",
                "text": "system"
              },
              {
                "type": "kw",
                "text": "shall"
              },
              {
                "type": "tok",
                "name": "verb",
                "kind": "action_verb",
                "values": [
                  "notify",
                  "alert",
                  "inform",
                  "warn"
                ]
              },
              {
                "type": "ref",
                "name": "target",
                "kind": "stakeholder"
              },
              {
                "type": "opt",
                "inner": [
                  {
                    "type": "kw",
                    "text": "using"
                  },
                  {
                    "type": "str",
                    "name": "mechanism"
                  }
                ]
              }
            ]
          },
          {
            "name": "Shall_persistence",
            "node_name": "action_clause",
            "clause_kind": "system_shall",
            "required": true,
            "repeatable": false,
            "order": 5,
            "pattern": [
              {
                "type": "kw",
                "text": "system"
              },
              {
                "type": "kw",
                "text": "shall"
              },
              {
                "type": "tok",
                "name": "verb",
                "kind": "action_verb",
                "values": [
                  "log",
                  "store",
                  "save",
                  "archive",
                  "retain"
                ]
              },
              {
                "type": "str",
                "name": "object"
              },
              {
                "type": "kw",
                "text": "of"
              },
              {
                "type": "ref",
                "name": "target",
                "kind": "stakeholder"
              },
              {
                "type": "opt",
                "inner": [
                  {
                    "type": "kw",
                    "text": "using"
                  },
                  {
                    "type": "str",
                    "name": "mechanism"
                  }
                ]
              }
            ]
          },
          {
            "name": "Shall_restriction",
            "node_name": "action_clause",
            "clause_kind": "system_shall",
            "required": true,
            "repeatable": false,
            "order": 5,
            "pattern": [
              {
                "type": "kw",
                "text": "system"
              },
              {
                "type": "kw",
                "text": "shall"
              },
              {
                "type": "tok",
                "name": "verb",
                "kind": "action_verb",
                "values": [
                  "restrict",
                  "limit",
                  "anonymize",
                  "mask",
                  "delete",
                  "remove"
                ]
              },
              {
                "type": "str",
                "name": "object"
              },
              {
                "type": "kw",
                "text": "of"
              },
              {
                "type": "ref",
                "name": "target",
                "kind": "stakeholder"
              }
            ]
          },
          {
            "name": "Stakeholders",
            "node_name": "stakeholders_clause",
            "clause_kind": "stakeholders",
            "required": true,
            "repeatable": false,
            "order": 6,
            "pattern": [
              {
                "type": "kw",
                "text": "stakeholders"
              },
              {
                "type": "list",
                "name": "actors",
                "separator": ",",
                "item": {
                  "type": "ref",
                  "kind": "stakeholder"
                }
              }
            ]
          },
          {
            "name": "Priority",
            "node_name": "priority_clause",
            "clause_kind": "priority",
            "required": false,
            "repeatable": false,
            "order": 7,
            "pattern": [
              {
                "type": "kw",
                "text": "priority"
              },
              {
                "type": "enum",
                "name": "level",
                "values": [
                  "low",
                  "medium",
                  "high",
                  "critical"
                ]
              }
            ]
          },
          {
            "name": "Retention",
            "node_name": "retention_clause",
            "clause_kind": "retention",
            "required": false,
            "repeatable": false,
            "order": 8,
            "pattern": [
              {
                "type": "kw",
                "text": "retention"
              },
              {
                "type": "str",
                "name": "value"
              }
            ]
          },
          {
            "name": "Access",
            "node_name": "access_clause",
            "clause_kind": "access",
            "required": false,
            "repeatable": false,
            "order": 9,
            "pattern": [
              {
                "type": "kw",
                "text": "access"
              },
              {
                "type": "str",
                "name": "value"
              }
            ]
          },
          {
            "name": "LinkedTo",
            "node_name": "linked_to_clause",
            "clause_kind": "linked_to",
            "required": false,
            "repeatable": true,
            "order": 10,
            "pattern": [
              {
                "type": "kw",
                "text": "linked_to"
              },
              {
                "type": "str",
                "name": "target"
              }
            ]
          }
        ]
      }
    },
    {
      "kind": "assignment",
      "node_name": "assignment_decl",
      "keyword": "assignment",
      "header": [
        {
          "type": "kw",
          "text": "assignment"
        },
        {
          "type": "ref",
          "name": "requirement",
          "kind": "requirement"
        }
      ],
      "body": {
        "lines": [
          {
            "name": "AssignmentEntry",
            "node_name": "assignment_entry",
            "required": true,
            "repeatable": true,
            "order": 1,
            "pattern": [
              {
                "type": "list",
                "name": "stakeholders",
                "separator": ",",
                "item": {
                  "type": "ref",
                  "kind": "stakeholder"
                }
              },
              {
                "type": "kw",
                "text": "->"
              },
              {
                "type": "list",
                "name": "values",
                "separator": ",",
                "item": {
                  "type": "ref",
                  "kind": "value"
                }
              }
            ]
          }
        ]
      }
    }
  ],
  "actions": [
    {
      "kind": "monitoring",
      "form": "object_of_target",
      "allows_mechanism": true,
      "description": "Observes or captures information about a target.",
      "template": "<verb> <object> of <target> [using <mechanism>]",
      "verbs": [
        "track",
        "monitor",
        "observe",
        "detect",
        "measure",
        "record",
        "collect",
        "capture"
      ]
    },
    {
      "kind": "notification",
      "form": "direct_target",
      "allows_mechanism": true,
      "description": "Informs or warns a target directly.",
      "template": "<verb> <target> [using <mechanism>]",
      "verbs": [
        "notify",
        "alert",
        "inform",
        "warn"
      ]
    },
    {
      "kind": "persistence",
      "form": "object_of_target",
      "allows_mechanism": true,
      "description": "Persists information about a target.",
      "template": "<verb> <object> of <target> [using <mechanism>]",
      "verbs": [
        "log",
        "store",
        "save",
        "archive",
        "retain"
      ]
    },
    {
      "kind": "restriction",
      "form": "object_of_target",
      "allows_mechanism": false,
      "description": "Constrains, removes, or obscures information about a target.",
      "template": "<verb> <object> of <target>",
      "verbs": [
        "restrict",
        "limit",
        "anonymize",
        "mask",
        "delete",
        "remove"
      ]
    }
  ]
};
