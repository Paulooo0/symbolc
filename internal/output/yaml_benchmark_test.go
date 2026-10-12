package output

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Paulooo0/symbolc/internal/analyzer"
	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
)

var benchmarkData []byte

func benchmarkConfig() config.Config {
	baseDir, _ := filepath.Abs(filepath.Join("..", ".."))
	return config.Config{
		BaseDir: baseDir,
		Source: config.SourceConfig{
			Roots: []string{"."},
		},
		Contracts: config.ContractsConfig{
			Include: []string{"internal/**"},
			Exclude: []string{"**/*_test.go", "**/generated/**"},
		},
	}
}

func benchmarkGraph(b *testing.B, cfg config.Config) *graph.Graph {
	b.Helper()
	g, err := analyzer.NewAnalyzer(cfg).Analyze(context.Background())
	if err != nil {
		b.Fatal(err)
	}
	return g
}

func BenchmarkMarshalYAML(b *testing.B) {
	cfg := benchmarkConfig()
	g := benchmarkGraph(b, cfg)

	data, err := MarshalYAML(g, cfg)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkData, err = MarshalYAML(g, cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzeAndMarshalYAML(b *testing.B) {
	cfg := benchmarkConfig()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g, err := analyzer.NewAnalyzer(cfg).Analyze(context.Background())
		if err != nil {
			b.Fatal(err)
		}
		benchmarkData, err = MarshalYAML(g, cfg)
		if err != nil {
			b.Fatal(err)
		}
		if i == 0 {
			b.SetBytes(int64(len(benchmarkData)))
		}
	}
}
