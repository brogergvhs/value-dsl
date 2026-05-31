; AUTO-GENERATED from generated/grammar.json by scripts/generate-highlights.mjs
(comment) @comment
(string_literal) @string
["," "->" "="] @operator
["access" "assignment" "if" "linked_to" "of" "priority" "requirement" "retention" "shall" "stakeholder" "stakeholders" "system" "using" "value" "when" "where" "while"] @keyword
(when_clause verb: (when_verb) @function.builtin)
(when_clause verb: (identifier) @function.builtin)
(priority_clause level: (identifier) @constant)
(targeted_action verb: (identifier) @function.builtin)
(direct_action verb: (identifier) @function.builtin)