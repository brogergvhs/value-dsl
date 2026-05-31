package grammar

// Kind enumerations for declaration, clause, and action vocabulary.

type RequirementClauseKind string

const (
	RequirementClauseWhile        RequirementClauseKind = "while"
	RequirementClauseWhen         RequirementClauseKind = "when"
	RequirementClauseIf           RequirementClauseKind = "if"
	RequirementClauseWhere        RequirementClauseKind = "where"
	RequirementClauseSystemShall  RequirementClauseKind = "system_shall"
	RequirementClauseStakeholders RequirementClauseKind = "stakeholders"
	RequirementClausePriority     RequirementClauseKind = "priority"
	RequirementClauseRetention    RequirementClauseKind = "retention"
	RequirementClauseAccess       RequirementClauseKind = "access"
	RequirementClauseLinkedTo     RequirementClauseKind = "linked_to"
)

type DeclarationKind string

const (
	DeclarationKindStakeholder DeclarationKind = "stakeholder"
	DeclarationKindValue       DeclarationKind = "value"
	DeclarationKindRequirement DeclarationKind = "requirement"
	DeclarationKindAssignment  DeclarationKind = "assignment"
)

type ActionKind string

const (
	ActionKindMonitoring   ActionKind = "monitoring"
	ActionKindNotification ActionKind = "notification"
	ActionKindPersistence  ActionKind = "persistence"
	ActionKindRestriction  ActionKind = "restriction"
)

// TokenKind classifies leaf tokens that are not plain keywords or
// identifiers. Closed sets (WhenVerb, ActionVerb) carry their members in
// the Tok matcher so the grammar file stays self-contained.
type TokenKind string

const (
	TokKindIdent      TokenKind = "ident"
	TokKindNumber     TokenKind = "number"
	TokKindString     TokenKind = "string"
	TokKindWhenVerb   TokenKind = "when_verb"
	TokKindActionVerb TokenKind = "action_verb"
)

// Matcher is the sum type for pattern primitives.
// Kw, Ident, Tok, Ref, Enum, List, Opt, Str, Free.
type Matcher interface{ matcher() }

// Kw matches an exact keyword literal.
type Kw struct{ Text string }

// Ident captures a raw identifier at the current position.
type Ident struct{ Name string }

// Tok captures a token of a given kind.
// Set enumerates members for closed kinds (e.g. WhenVerb).
// Set is ignored for open kinds like Ident/Number.
type Tok struct {
	Kind TokenKind
	Name string
	Set  []string
}

// Ref captures an identifier that must resolve to a declaration of Kind.
type Ref struct {
	Name string
	Kind DeclarationKind
}

// Enum captures one of a fixed set of literal values (e.g. priority levels).
type Enum struct {
	Name   string
	Values []string
}

// List matches one or more Items separated by Separator (typically ",").
type List struct {
	Name      string
	Item      Matcher
	Separator string
}

// Opt matches a subsequence zero or one time.
type Opt struct {
	Inner []Matcher
}

// Str captures an identifier or a quoted string literal at the current position.
// Use in place of Ident where multi-word text is permitted (mechanism, object, traceability).
type Str struct{ Name string }

// Free captures the remainder of the line as raw text. Use sparingly —
// it defers shape checking to a custom lowering step.
type Free struct{ Name string }

func (Kw) matcher()    {}
func (Ident) matcher() {}
func (Tok) matcher()   {}
func (Ref) matcher()   {}
func (Enum) matcher()  {}
func (List) matcher()  {}
func (Opt) matcher()   {}
func (Str) matcher()   {}
func (Free) matcher()  {}

// SemanticRole tags a pattern field with its role in semantic interpretation.
type SemanticRole string

const (
	SemanticRoleAssignmentSource SemanticRole = "assignment_source"
	SemanticRoleAssignmentTarget SemanticRole = "assignment_target"
)

// SemanticBinding maps a pattern capture field to a semantic role.
type SemanticBinding struct {
	Role  SemanticRole
	Field string
}

// ExpansionMode controls how multi-valued captures expand into semantic objects.
type ExpansionMode string

const (
	ExpansionCartesian ExpansionMode = "cartesian"
)

// LineRule describes one line shape inside a block. ClauseKind is an
// optional semantic tag used by lowering and validation when the line
// corresponds to a known requirement clause kind.
type LineRule struct {
	Name       string
	ClauseKind RequirementClauseKind
	Pattern    []Matcher
	Required   bool
	Repeatable bool
	Order      int
	Semantics  []SemanticBinding
	Expansion  ExpansionMode
}

// Block groups the line rules that may appear in a declaration body.
type Block struct {
	Lines []LineRule
}

// Decl is a top-level declaration: keyword + header pattern + optional body.
type Decl struct {
	Kind       DeclarationKind
	Keyword    string
	Header     []Matcher
	Body       *Block
	DenseGroup bool
}

// Spec is the whole language as data.
type Spec struct {
	Identifier    string
	NumberPattern string
	StringPattern string
	LineComment   string
	Declarations  []Decl
}
