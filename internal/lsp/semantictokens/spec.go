package semantictokens

type TokenType string

const (
	TokenTypeType       TokenType = "type"
	TokenTypeEnumMember TokenType = "enumMember"
	TokenTypeFunction   TokenType = "function"
)

type TokenModifier string

const (
	TokenModifierDeclaration    TokenModifier = "declaration"
	TokenModifierDefaultLibrary TokenModifier = "defaultLibrary"
)

type Role string

const (
	RoleStakeholderDeclaration Role = "stakeholder_declaration"
	RoleStakeholderReference   Role = "stakeholder_reference"
	RoleValueDeclaration       Role = "value_declaration"
	RoleValueReference         Role = "value_reference"
	RoleRequirementDeclaration Role = "requirement_declaration"
	RoleRequirementReference   Role = "requirement_reference"
	RoleBuiltinValue           Role = "builtin_value"
)

type Spec struct {
	Role      Role
	Type      TokenType
	Modifiers []TokenModifier
}

var specs = []Spec{
	{Role: RoleStakeholderDeclaration, Type: TokenTypeType, Modifiers: []TokenModifier{TokenModifierDeclaration}},
	{Role: RoleStakeholderReference, Type: TokenTypeType},
	{Role: RoleValueDeclaration, Type: TokenTypeEnumMember, Modifiers: []TokenModifier{TokenModifierDeclaration}},
	{Role: RoleValueReference, Type: TokenTypeEnumMember},
	{Role: RoleRequirementDeclaration, Type: TokenTypeFunction, Modifiers: []TokenModifier{TokenModifierDeclaration}},
	{Role: RoleRequirementReference, Type: TokenTypeFunction},
	{Role: RoleBuiltinValue, Type: TokenTypeEnumMember, Modifiers: []TokenModifier{TokenModifierDefaultLibrary}},
}

type Match struct {
	Line        int
	StartColumn int
	EndColumn   int
	Role        Role
}

func LegendTypes() []string {
	out := make([]string, 0, len(specs))
	seen := make(map[TokenType]struct{}, len(specs))
	for _, spec := range specs {
		if _, ok := seen[spec.Type]; ok {
			continue
		}
		seen[spec.Type] = struct{}{}
		out = append(out, string(spec.Type))
	}
	return out
}

func LegendModifiers() []string {
	out := make([]string, 0, 4)
	seen := make(map[TokenModifier]struct{}, 4)
	for _, spec := range specs {
		for _, modifier := range spec.Modifiers {
			if _, ok := seen[modifier]; ok {
				continue
			}
			seen[modifier] = struct{}{}
			out = append(out, string(modifier))
		}
	}
	return out
}

func SpecForRole(role Role) (Spec, bool) {
	for _, spec := range specs {
		if spec.Role == role {
			return spec, true
		}
	}
	return Spec{}, false
}
