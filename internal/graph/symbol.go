package graph

import "fmt"

type SymbolID string

type SymbolKind string

const (
	KindUnknown   SymbolKind = "unknown"
	KindPackage   SymbolKind = "package"
	KindStruct    SymbolKind = "struct"
	KindInterface SymbolKind = "interface"
	KindFunction  SymbolKind = "function"
	KindMethod    SymbolKind = "method"
	KindConstant  SymbolKind = "constant"
	KindVariable  SymbolKind = "variable"
)

type Symbol struct {
	ID         SymbolID
	Name       string
	Package    Package
	Kind       SymbolKind
	Source     Source
	Methods    []SymbolID
	References []SymbolID
}

type Package struct {
	Name       string
	ImportPath string
}

type Source struct {
	File   string
	Line   int
	Column int
}

func NewSymbolID(importPath, name string) SymbolID {
	return SymbolID(fmt.Sprintf("%s.%s", importPath, name))
}
