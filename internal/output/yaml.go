package output

import (
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
	"gopkg.in/yaml.v3"
)

const version = 1

type document struct {
	Version       int            `yaml:"version"`
	Packages      []packageEntry `yaml:"packages"`
	Symbols       []symbolEntry  `yaml:"symbols"`
	Relationships []relationship `yaml:"relationships,omitempty"`
}

type packageEntry struct {
	Name       string `yaml:"name"`
	ImportPath string `yaml:"import_path"`
}

type symbolEntry struct {
	ID      graph.SymbolID   `yaml:"id"`
	Name    string           `yaml:"name"`
	Package string           `yaml:"package"`
	Kind    graph.SymbolKind `yaml:"kind"`
	Source  graph.Source     `yaml:"source"`
}

type relationship struct {
	From graph.SymbolID `yaml:"from"`
	To   graph.SymbolID `yaml:"to"`
	Kind string         `yaml:"kind"`
}

const methodRelationship = "method"

func MarshalYAML(g *graph.Graph, cfg config.Config) ([]byte, error) {
	doc := document{Version: version}

	ids := make([]graph.SymbolID, 0, len(g.Symbols))
	for id := range g.Symbols {
		if !included(g.Symbols[id], cfg) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	packages := make(map[string]packageEntry)
	for _, id := range ids {
		symbol := g.Symbols[id]
		doc.Symbols = append(doc.Symbols, symbolEntry{
			ID:      symbol.ID,
			Name:    symbol.Name,
			Package: symbol.Package.ImportPath,
			Kind:    symbol.Kind,
			Source:  symbol.Source,
		})

		packages[symbol.Package.ImportPath] = packageEntry{
			Name:       symbol.Package.Name,
			ImportPath: symbol.Package.ImportPath,
		}

		for _, methodID := range symbol.Methods {
			if _, ok := g.Symbols[methodID]; !ok || !included(g.Symbols[methodID], cfg) {
				continue
			}
			doc.Relationships = append(doc.Relationships, relationship{
				From: symbol.ID,
				To:   methodID,
				Kind: methodRelationship,
			})
		}
	}

	for _, pkg := range packages {
		doc.Packages = append(doc.Packages, pkg)
	}
	sort.Slice(doc.Packages, func(i, j int) bool {
		return doc.Packages[i].ImportPath < doc.Packages[j].ImportPath
	})
	sort.Slice(doc.Relationships, func(i, j int) bool {
		if doc.Relationships[i].From != doc.Relationships[j].From {
			return doc.Relationships[i].From < doc.Relationships[j].From
		}
		if doc.Relationships[i].To != doc.Relationships[j].To {
			return doc.Relationships[i].To < doc.Relationships[j].To
		}
		return doc.Relationships[i].Kind < doc.Relationships[j].Kind
	})

	return yaml.Marshal(doc)
}

func included(symbol *graph.Symbol, cfg config.Config) bool {
	if symbol.Source.File == "" {
		return true
	}

	paths := sourcePaths(symbol.Source.File, cfg.Source.Roots)
	if len(cfg.Contracts.Include) > 0 && !matchesAny(paths, cfg.Contracts.Include) {
		return false
	}

	return !matchesAny(paths, cfg.Contracts.Exclude)
}

func sourcePaths(file string, roots []string) []string {
	paths := []string{filepath.ToSlash(filepath.Clean(file))}
	absFile, err := filepath.Abs(file)
	if err != nil {
		return paths
	}

	if cwd, err := filepath.Abs("."); err == nil {
		if relative, err := filepath.Rel(cwd, absFile); err == nil && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			paths = append(paths, filepath.ToSlash(relative))
		}
	}

	for _, root := range roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		relative, err := filepath.Rel(absRoot, absFile)
		if err == nil && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			paths = append(paths, filepath.ToSlash(relative))
		}
	}

	return paths
}

func matchesAny(paths, patterns []string) bool {
	for _, file := range paths {
		for _, pattern := range patterns {
			if matchPath(file, pattern) {
				return true
			}
		}
	}
	return false
}

func matchPath(file, pattern string) bool {
	fileParts := strings.Split(filepath.ToSlash(filepath.Clean(file)), "/")
	patternParts := strings.Split(filepath.ToSlash(filepath.Clean(pattern)), "/")
	return matchPathParts(fileParts, patternParts)
}

func matchPathParts(file, pattern []string) bool {
	if len(pattern) == 0 {
		return len(file) == 0
	}
	if pattern[0] == "**" {
		return matchPathParts(file, pattern[1:]) ||
			(len(file) > 0 && matchPathParts(file[1:], pattern))
	}
	if len(file) == 0 {
		return false
	}

	matched, err := path.Match(pattern[0], file[0])
	return err == nil && matched && matchPathParts(file[1:], pattern[1:])
}
