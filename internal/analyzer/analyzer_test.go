package analyzer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_Analyze(t *testing.T) {
	root := filepath.Join("testdata", "basic")

	cfg := config.Config{
		Source: config.SourceConfig{
			Roots: []string{root},
		},
	}

	analyzer := NewAnalyzer(cfg)

	got, err := analyzer.Analyze(context.Background())
	require.NoError(t, err)

	t.Run("discovers_symbols", func(t *testing.T) {
		require.Len(t, got.Symbols, 9)

		expected := map[graph.SymbolID]graph.SymbolKind{
			"example.com/testproject.User":                      graph.KindStruct,
			"example.com/testproject.CreateUserInput":           graph.KindStruct,
			"example.com/testproject.UserService":               graph.KindStruct,
			"example.com/testproject.UserRepository":            graph.KindInterface,
			"example.com/testproject.ErrUserNotFound":           graph.KindVariable,
			"example.com/testproject.MaxRetries":                graph.KindConstant,
			"example.com/testproject.UserRepository.CreateUser": graph.KindMethod,
			"example.com/testproject.UserService.CreateUser":    graph.KindMethod,
			"example.com/testproject.Foo":                       graph.KindFunction,
		}

		for id, expectedKind := range expected {
			t.Run(string(id), func(t *testing.T) {
				symbol, ok := got.Symbols[id]
				require.True(t, ok, "symbol %s not found", id)
				require.Equal(t, expectedKind, symbol.Kind)
			})
		}
	})

	t.Run("ignores_local_variables", func(t *testing.T) {
		require.NotContains(
			t,
			got.Symbols,
			graph.SymbolID("example.com/testproject.user"),
		)
	})

	t.Run("associates_methods_with_receivers", func(t *testing.T) {
		expected := map[graph.SymbolID][]graph.SymbolID{
			"example.com/testproject.UserService": {
				"example.com/testproject.UserService.CreateUser",
			},
			"example.com/testproject.UserRepository": {
				"example.com/testproject.UserRepository.CreateUser",
			},
		}

		for receiverID, expectedMethods := range expected {
			t.Run(string(receiverID), func(t *testing.T) {
				receiver, ok := got.Symbols[receiverID]
				require.True(t, ok, "receiver %s not found", receiverID)

				require.ElementsMatch(
					t,
					expectedMethods,
					receiver.Methods,
				)
			})
		}
	})
}
