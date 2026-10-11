package output

import (
	"testing"

	"github.com/Paulooo0/symbolc/internal/graph"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshalYAML(t *testing.T) {
	g := graph.NewGraph()
	receiver := &graph.Symbol{
		ID:   "example.com/testproject.User",
		Name: "User",
		Package: graph.Package{
			Name:       "main",
			ImportPath: "example.com/testproject",
		},
		Kind: graph.KindStruct,
	}
	method := &graph.Symbol{
		ID:   "example.com/testproject.User.String",
		Name: "String",
		Package: graph.Package{
			Name:       "main",
			ImportPath: "example.com/testproject",
		},
		Kind: graph.KindMethod,
	}
	packageSymbol := &graph.Symbol{
		ID:   "example.com/testproject",
		Name: "main",
		Package: graph.Package{
			Name:       "main",
			ImportPath: "example.com/testproject",
		},
		Kind: graph.KindPackage,
	}

	g.AddSymbol(receiver)
	g.AddSymbol(method)
	g.AddSymbol(packageSymbol)
	g.AddMethod(receiver.ID, method.ID)

	first, err := MarshalYAML(g)
	require.NoError(t, err)

	second, err := MarshalYAML(g)
	require.NoError(t, err)
	require.Equal(t, first, second)

	var got document
	require.NoError(t, yaml.Unmarshal(first, &got))
	require.Equal(t, 1, got.Version)
	require.Len(t, got.Packages, 1)
	require.Len(t, got.Symbols, 3)
	require.Equal(t, []relationship{
		{
			From: receiver.ID,
			To:   method.ID,
			Kind: methodRelationship,
		},
	}, got.Relationships)
}

func TestMarshalYAMLSkipsMissingMethodSymbols(t *testing.T) {
	g := graph.NewGraph()
	receiver := &graph.Symbol{ID: "User", Methods: []graph.SymbolID{"missing"}}
	g.AddSymbol(receiver)

	data, err := MarshalYAML(g)
	require.NoError(t, err)
	require.NotContains(t, string(data), "missing")
}
