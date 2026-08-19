package analyzer

import (
	"context"
	"fmt"
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

	for _, s := range got.Symbols {
		fmt.Printf("\n%+v\n", s)
	}

	require.Len(t, got.Symbols, 6)

	tests := map[graph.SymbolID]graph.SymbolKind{
		"example.com/testproject.User":            graph.KindStruct,
		"example.com/testproject.CreateUserInput": graph.KindStruct,
		"example.com/testproject.UserService":     graph.KindStruct,
		"example.com/testproject.UserRepository":  graph.KindInterface,
		"example.com/testproject.ErrUserNotFound": graph.KindVariable,
		"example.com/testproject.MaxRetries":      graph.KindConstant,
	}

	for id, expectedKind := range tests {
		symbol, ok := got.Symbols[id]

		require.True(t, ok, "symbol %s not found", id)
		require.Equal(t, expectedKind, symbol.Kind)
	}

	require.NotContains(t, got.Symbols, graph.SymbolID(
		"example.com/testproject.user",
	))
}
