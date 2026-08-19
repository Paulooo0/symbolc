package graph

type DependencyKind string

const (
	DependencyReference  DependencyKind = "reference"
	DependencyCall       DependencyKind = "call"
	DependencyEmbed      DependencyKind = "embed"
	DependencyImplements DependencyKind = "implements"
)

type Dependecy struct {
	From SymbolID
	To   SymbolID
	Kind DependencyKind
}
