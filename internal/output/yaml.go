package output

import (
	"sort"

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

func MarshalYAML(g *graph.Graph) ([]byte, error) {
	doc := document{Version: version}

	ids := make([]graph.SymbolID, 0, len(g.Symbols))
	for id := range g.Symbols {
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

		if symbol.Kind == graph.KindPackage {
			packages[symbol.Package.ImportPath] = packageEntry{
				Name:       symbol.Package.Name,
				ImportPath: symbol.Package.ImportPath,
			}
		}

		for _, methodID := range symbol.Methods {
			if _, ok := g.Symbols[methodID]; !ok {
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
