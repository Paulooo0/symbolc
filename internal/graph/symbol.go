package graph

import (
	"fmt"
	"go/types"
)

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

func NewSymbolID(obj types.Object) SymbolID {
	if fn, ok := obj.(*types.Func); ok {
		if signature, ok := fn.Type().(*types.Signature); ok {
			if recv := signature.Recv(); recv != nil {
				return SymbolID(fmt.Sprintf(
					"%s.%s.%s",
					obj.Pkg().Path(),
					receiverName(recv.Type()),
					fn.Name(),
				))
			}
		}
	}

	if obj.Pkg() == nil {
		return SymbolID(obj.Name())
	}

	return SymbolID(fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()))
}

func receiverName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Name()

	case *types.Alias:
		return receiverName(t.Rhs())

	case *types.Pointer:
		return receiverName(t.Elem())

	default:
		return ""
	}
}
