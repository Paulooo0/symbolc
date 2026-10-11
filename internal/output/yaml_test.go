package output

import (
	"testing"

	"github.com/Paulooo0/symbolc/internal/config"
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

	first, err := MarshalYAML(g, config.Config{})
	require.NoError(t, err)

	second, err := MarshalYAML(g, config.Config{})
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

	data, err := MarshalYAML(g, config.Config{})
	require.NoError(t, err)
	require.NotContains(t, string(data), "missing")
}

func TestMarshalYAMLFiltersBySourcePath(t *testing.T) {
	g := graph.NewGraph()
	for _, symbol := range []*graph.Symbol{
		{ID: "included", Package: graph.Package{ImportPath: "example.com/test"}, Source: graph.Source{File: "/project/internal/service/service.go"}},
		{ID: "test", Package: graph.Package{ImportPath: "example.com/test"}, Source: graph.Source{File: "/project/internal/service/service_test.go"}},
		{ID: "generated", Package: graph.Package{ImportPath: "example.com/test"}, Source: graph.Source{File: "/project/internal/generated/code.go"}},
	} {
		g.AddSymbol(symbol)
	}

	cfg := config.Config{
		Source: config.SourceConfig{Roots: []string{"/project"}},
		Contracts: config.ContractsConfig{
			Include: []string{"internal/**"},
			Exclude: []string{"**/*_test.go", "**/generated/**"},
		},
	}

	data, err := MarshalYAML(g, cfg)
	require.NoError(t, err)

	var got document
	require.NoError(t, yaml.Unmarshal(data, &got))
	require.Len(t, got.Symbols, 1)
	require.Equal(t, graph.SymbolID("included"), got.Symbols[0].ID)
}
