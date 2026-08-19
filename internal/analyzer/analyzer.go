package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"go/types"

	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
	"golang.org/x/tools/go/packages"
)

type Analyzer interface {
	Analyze(ctx context.Context) (*graph.Graph, error)
}

type analyzer struct {
	cfg config.Config
}

func NewAnalyzer(cfg config.Config) Analyzer {
	return &analyzer{
		cfg: cfg,
	}
}

func (a *analyzer) Analyze(ctx context.Context) (*graph.Graph, error) {
	pkgs, err := a.loadPackages()
	if err != nil {
		return nil, err
	}

	g := graph.NewGraph()

	for _, pkg := range pkgs {
		a.analyzePackage(g, pkg)
	}

	return g, nil
}

func (a *analyzer) loadPackages() (
	[]*packages.Package, error,
) {
	// Try to load packages from each root directory
	var allPkgs []*packages.Package

	for _, root := range a.cfg.Source.Roots {
		cfg := &packages.Config{
			Mode: packages.NeedName |
				packages.NeedCompiledGoFiles |
				packages.NeedImports |
				packages.NeedDeps |
				packages.NeedTypes |
				packages.NeedTypesInfo |
				packages.NeedSyntax,
			Dir: root,
		}

		// Load packages from this directory
		pkgs, err := packages.Load(cfg, ".")
		if err != nil {
			return nil, fmt.Errorf("analyzer: failed load packages from %s: %w", root, err)
		}

		if n := packages.PrintErrors(pkgs); n > 0 {
			return nil, fmt.Errorf("analyzer: packages contain %d error(s)", n)
		}

		allPkgs = append(allPkgs, pkgs...)
	}

	return allPkgs, nil
}

func (a *analyzer) analyzePackage(g *graph.Graph, pkg *packages.Package) {
	scope := pkg.Types.Scope()

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		switch obj := obj.(type) {
		case *types.TypeName:
			a.analyzeTypeName(g, pkg, obj)
		case *types.Const:
			a.analyzeConst(g, pkg, obj)
		case *types.Var:
			a.analyzeVar(g, pkg, obj)
		case *types.Func:
			a.analyzeFunc(g, pkg, obj)
		}
	}
}

func (a *analyzer) analyzeTypeName(g *graph.Graph, pkg *packages.Package, typeName *types.TypeName) {
	named, ok := typeName.Type().(*types.Named)
	if !ok {
		return
	}

	g.AddSymbol(symbolFromNamedType(pkg, typeName, named))
}

func (a *analyzer) analyzeConst(g *graph.Graph, pkg *packages.Package, constObj *types.Const) {
	g.AddSymbol(symbolFromObj(pkg, constObj, graph.KindConstant))
}

func (a *analyzer) analyzeVar(g *graph.Graph, pkg *packages.Package, varObj *types.Var) {
	g.AddSymbol(symbolFromObj(pkg, varObj, graph.KindVariable))
}

func (a *analyzer) analyzeFunc(g *graph.Graph, pkg *packages.Package, funcObj *types.Func) {
	g.AddSymbol(symbolFromObj(pkg, funcObj, graph.KindFunction))
}

func symbolFromNamedType(
	pkg *packages.Package,
	typeName *types.TypeName,
	named *types.Named,
) graph.Symbol {
	kind := graph.KindUnknown

	switch named.Underlying().(type) {
	case *types.Struct:
		kind = graph.KindStruct
	case *types.Interface:
		kind = graph.KindInterface
	}

	return graph.Symbol{
		ID:   graph.SymbolID(fmt.Sprintf("%s.%s", pkg.PkgPath, typeName.Name())),
		Name: typeName.Name(),
		Package: graph.Package{
			Name:       pkg.Name,
			ImportPath: pkg.PkgPath,
		},
		Kind:   kind,
		Source: sourceOf(pkg, typeName.Pos()),
	}
}

func symbolFromObj(
	pkg *packages.Package,
	obj types.Object,
	kind graph.SymbolKind,
) graph.Symbol {
	return graph.Symbol{
		ID: graph.SymbolID(
			fmt.Sprintf("%s.%s", pkg.PkgPath, obj.Name()),
		),
		Name: obj.Name(),
		Package: graph.Package{
			Name:       pkg.Name,
			ImportPath: pkg.PkgPath,
		},
		Kind:   kind,
		Source: sourceOf(pkg, obj.Pos()),
	}
}

func sourceOf(
	pkg *packages.Package,
	pos token.Pos,
) graph.Source {
	position := pkg.Fset.Position(pos)

	return graph.Source{
		File:   position.Filename,
		Line:   position.Line,
		Column: position.Column,
	}
}
