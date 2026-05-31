; AUTO-GENERATED from generated/grammar.json by scripts/generate-locals.mjs

(source_file) @local.scope

; Definitions
(requirement_decl id: (identifier) @local.definition)
(stakeholder_decl names: (identifier) @local.definition)
(value_decl name: (identifier) @local.definition)

; References
(action_clause (action_expression (direct_action target: (identifier) @local.reference)))
(action_clause (action_expression (targeted_action target: (identifier) @local.reference)))
(assignment_decl requirement: (identifier) @local.reference)
(if_clause actor: (identifier) @local.reference)
(when_clause actor: (identifier) @local.reference)
(while_clause actor: (identifier) @local.reference)
