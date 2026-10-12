package output

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Paulooo0/symbolc/internal/analyzer"
	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
)

var benchmarkFiles map[string][]byte

func benchmarkConfig() config.Config {
	baseDir, _ := filepath.Abs(filepath.Join("..", ".."))
	return config.Config{
		BaseDir: baseDir,
		Source:  config.SourceConfig{Roots: []string{"."}},
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

func benchmarkSize(files map[string][]byte) int64 {
	var size int64
	for _, data := range files {
		size += int64(len(data))
	}
	return size
}

func BenchmarkMaterialize(b *testing.B) {
	cfg := benchmarkConfig()
	g := benchmarkGraph(b, cfg)
	files, err := Materialize(g, cfg)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(benchmarkSize(files))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkFiles, err = Materialize(g, cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzeAndMaterialize(b *testing.B) {
	cfg := benchmarkConfig()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g := benchmarkGraph(b, cfg)
		var err error
		benchmarkFiles, err = Materialize(g, cfg)
		if err != nil {
			b.Fatal(err)
		}
		if i == 0 {
			b.SetBytes(benchmarkSize(benchmarkFiles))
		}
	}
}
