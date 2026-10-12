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

type artifactIndex struct {
	Version         int           `yaml:"version"`
	Executables     []artifactRef `yaml:"executables,omitempty"`
	FluxEntrypoints []artifactRef `yaml:"flux_entrypoints,omitempty"`
	Contracts       []artifactRef `yaml:"contracts,omitempty"`
}

type artifactRef struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type contract struct {
	Version int              `yaml:"version"`
	ID      graph.SymbolID   `yaml:"id"`
	Name    string           `yaml:"name"`
	Kind    graph.SymbolKind `yaml:"kind"`
	Package packageRef       `yaml:"package"`
	Source  graph.Source     `yaml:"source"`
	Methods []graph.SymbolID `yaml:"methods,omitempty"`
}

type packageRef struct {
	Name       string `yaml:"name"`
	ImportPath string `yaml:"import_path"`
}

type fluxEntrypoint struct {
	Version  int            `yaml:"version"`
	ID       graph.SymbolID `yaml:"id"`
	Name     string         `yaml:"name"`
	Packages []fluxPackage  `yaml:"packages"`
}

type fluxPackage struct {
	ID      graph.SymbolID   `yaml:"id"`
	Name    string           `yaml:"name"`
	Imports []graph.SymbolID `yaml:"imports,omitempty"`
}

func Materialize(g *graph.Graph, cfg config.Config) (map[string][]byte, error) {
	selected := selectedSymbols(g, cfg)
	mainIDs := mainPackages(g)
	if len(mainIDs) == 0 {
		return materializeSession(g, cfg, selected, nil)
	}

	files := make(map[string][]byte)
	executables := make([]artifactRef, 0, len(mainIDs))
	for _, mainID := range mainIDs {
		scope := sessionSymbols(g, mainID, selected)
		roots := directEntrypoints(g, mainID, scope)
		session, err := materializeSession(g, cfg, scope, roots)
		if err != nil {
			return nil, err
		}
		prefix := executablePath(g, cfg, mainID)
		executables = append(executables, artifactRef{
			ID:   string(mainID),
			Path: filepath.ToSlash(filepath.Join(prefix, "index.yml")),
		})
		for path, data := range session {
			if prefix != "" {
				path = filepath.ToSlash(filepath.Join(prefix, path))
			}
			files[path] = data
		}
	}
	index, err := yaml.Marshal(artifactIndex{Version: version, Executables: executables})
	if err != nil {
		return nil, err
	}
	files["index.yml"] = index
	return files, nil
}

func materializeSession(g *graph.Graph, cfg config.Config, selected, roots []graph.SymbolID) (map[string][]byte, error) {
	contractPaths := artifactPaths(selected, g)
	files := make(map[string][]byte)
	contracts := make([]artifactRef, 0, len(selected))
	for _, id := range selected {
		symbol := g.Symbols[id]
		path := contractPaths[id]
		data, err := yaml.Marshal(contract{
			Version: version,
			ID:      symbol.ID,
			Name:    symbol.Name,
			Kind:    symbol.Kind,
			Package: packageRef{Name: symbol.Package.Name, ImportPath: symbol.Package.ImportPath},
			Source:  relativeSource(symbol.Source, cfg),
			Methods: selectedMethods(symbol, selected),
		})
		if err != nil {
			return nil, err
		}
		files[path] = data
		contracts = append(contracts, artifactRef{ID: string(id), Path: path})
	}

	entrypoints, err := materializeEntrypoints(g, roots, selected)
	if err != nil {
		return nil, err
	}
	for path, data := range entrypoints.files {
		files[path] = data
	}

	index, err := yaml.Marshal(artifactIndex{Version: version, FluxEntrypoints: entrypoints.refs, Contracts: contracts})
	if err != nil {
		return nil, err
	}
	files["index.yml"] = index
	return files, nil
}

func mainPackages(g *graph.Graph) []graph.SymbolID {
	ids := make([]graph.SymbolID, 0)
	for id, pkg := range g.Packages {
		if pkg.Main {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func sessionSymbols(g *graph.Graph, mainID graph.SymbolID, selected []graph.SymbolID) []graph.SymbolID {
	selectedSet := make(map[graph.SymbolID]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[id] = struct{}{}
	}

	packages := make(map[graph.SymbolID]bool)
	var visit func(graph.SymbolID)
	visit = func(id graph.SymbolID) {
		if packages[id] {
			return
		}
		packages[id] = true
		pkg, ok := g.Packages[id]
		if !ok {
			return
		}
		for _, imported := range pkg.Imports {
			if _, ok := selectedSet[imported]; ok {
				visit(imported)
			}
		}
	}
	visit(mainID)

	result := make([]graph.SymbolID, 0)
	for _, id := range selected {
		symbol := g.Symbols[id]
		if packages[graph.SymbolID(symbol.Package.ImportPath)] {
			result = append(result, id)
		}
	}
	return result
}

func directEntrypoints(g *graph.Graph, mainID graph.SymbolID, selected []graph.SymbolID) []graph.SymbolID {
	selectedSet := make(map[graph.SymbolID]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[id] = struct{}{}
	}
	result := make([]graph.SymbolID, 0)
	for _, imported := range g.Packages[mainID].Imports {
		if _, ok := g.Packages[imported]; ok {
			if _, ok := selectedSet[imported]; ok {
				result = append(result, imported)
			}
		}
	}
	return result
}

func executablePath(g *graph.Graph, cfg config.Config, mainID graph.SymbolID) string {
	symbol := g.Symbols[mainID]
	if symbol == nil {
		return safeName(string(mainID))
	}
	relative := relativeSource(symbol.Source, cfg).File
	if relative == "" || filepath.IsAbs(relative) {
		return safeName(string(mainID))
	}
	directory := filepath.ToSlash(filepath.Dir(relative))
	if directory == "." {
		return safeName(string(mainID))
	}
	return directory
}

func selectedSymbols(g *graph.Graph, cfg config.Config) []graph.SymbolID {
	ids := make([]graph.SymbolID, 0, len(g.Symbols))
	for id, symbol := range g.Symbols {
		if included(symbol, cfg) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func selectedMethods(symbol *graph.Symbol, selected []graph.SymbolID) []graph.SymbolID {
	allowed := make(map[graph.SymbolID]struct{}, len(selected))
	for _, id := range selected {
		allowed[id] = struct{}{}
	}

	methods := make([]graph.SymbolID, 0, len(symbol.Methods))
	for _, id := range symbol.Methods {
		if _, ok := allowed[id]; ok {
			methods = append(methods, id)
		}
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i] < methods[j] })
	return methods
}

func artifactPaths(ids []graph.SymbolID, g *graph.Graph) map[graph.SymbolID]string {
	paths := make(map[graph.SymbolID]string, len(ids))
	counts := make(map[string]int)
	for _, id := range ids {
		symbol := g.Symbols[id]
		candidate := filepath.ToSlash(filepath.Join(
			"contracts",
			safeName(path.Base(symbol.Package.ImportPath)),
			safeName(contractName(symbol)),
		)) + ".yml"
		paths[id] = candidate
		counts[candidate]++
	}

	for _, id := range ids {
		if counts[paths[id]] > 1 {
			paths[id] = filepath.ToSlash(filepath.Join("contracts", safeName(string(id)))) + ".yml"
		}
	}
	return paths
}

func contractName(symbol *graph.Symbol) string {
	name := strings.TrimPrefix(string(symbol.ID), symbol.Package.ImportPath+".")
	if name == string(symbol.ID) || name == "" {
		name = "package"
	}
	return name
}

func materializeEntrypoints(g *graph.Graph, roots, selected []graph.SymbolID) (struct {
	files map[string][]byte
	refs  []artifactRef
}, error) {
	selectedSet := make(map[graph.SymbolID]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[id] = struct{}{}
	}

	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	roots = uniqueIDs(roots)

	paths := make(map[graph.SymbolID]string, len(roots))
	counts := make(map[string]int)
	for _, id := range roots {
		candidate := filepath.ToSlash(filepath.Join("flux-entrypoints", safeName(g.Packages[id].Name))) + ".yml"
		paths[id] = candidate
		counts[candidate]++
	}

	files := make(map[string][]byte)
	refs := make([]artifactRef, 0, len(roots))
	for _, root := range roots {
		if counts[paths[root]] > 1 {
			paths[root] = filepath.ToSlash(filepath.Join("flux-entrypoints", safeName(string(root)))) + ".yml"
		}
		packages := reachablePackages(g, root, selectedSet)
		data, err := yaml.Marshal(fluxEntrypoint{
			Version:  version,
			ID:       root,
			Name:     g.Packages[root].Name,
			Packages: packages,
		})
		if err != nil {
			return struct {
				files map[string][]byte
				refs  []artifactRef
			}{}, err
		}
		files[paths[root]] = data
		refs = append(refs, artifactRef{ID: string(root), Path: paths[root]})
	}

	return struct {
		files map[string][]byte
		refs  []artifactRef
	}{files: files, refs: refs}, nil
}

func reachablePackages(g *graph.Graph, root graph.SymbolID, selected map[graph.SymbolID]struct{}) []fluxPackage {
	seen := make(map[graph.SymbolID]bool)
	var visit func(graph.SymbolID)
	visit = func(id graph.SymbolID) {
		if seen[id] {
			return
		}
		seen[id] = true
		pkg, ok := g.Packages[id]
		if !ok {
			return
		}
		for _, imported := range pkg.Imports {
			if _, ok := selected[imported]; ok {
				visit(imported)
			}
		}
	}
	visit(root)

	ids := make([]graph.SymbolID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	packages := make([]fluxPackage, 0, len(ids))
	for _, id := range ids {
		pkg := g.Packages[id]
		imports := make([]graph.SymbolID, 0)
		for _, imported := range pkg.Imports {
			if seen[imported] {
				imports = append(imports, imported)
			}
		}
		packages = append(packages, fluxPackage{ID: id, Name: pkg.Name, Imports: imports})
	}
	return packages
}

func uniqueIDs(ids []graph.SymbolID) []graph.SymbolID {
	result := ids[:0]
	for _, id := range ids {
		if len(result) == 0 || result[len(result)-1] != id {
			result = append(result, id)
		}
	}
	return result
}

func safeName(value string) string {
	value = strings.ReplaceAll(value, "..", "_")
	value = strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(value)
	if value == "" {
		return "unknown"
	}
	return value
}

func included(symbol *graph.Symbol, cfg config.Config) bool {
	if symbol.Source.File == "" {
		return true
	}

	paths := sourcePaths(symbol.Source.File, cfg)
	if len(cfg.Contracts.Include) > 0 && !matchesAny(paths, cfg.Contracts.Include) {
		return false
	}

	return !matchesAny(paths, cfg.Contracts.Exclude)
}

func sourcePaths(file string, cfg config.Config) []string {
	paths := []string{filepath.ToSlash(filepath.Clean(file))}
	absFile, err := filepath.Abs(file)
	if err != nil {
		return paths
	}

	baseDir := cfg.BaseDir
	if baseDir == "" {
		baseDir, _ = filepath.Abs(".")
	}
	if relative, err := filepath.Rel(baseDir, absFile); err == nil && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		paths = append(paths, filepath.ToSlash(relative))
	}

	for _, root := range cfg.Source.Roots {
		if baseDir != "" && !filepath.IsAbs(root) {
			root = filepath.Join(baseDir, root)
		}
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

func relativeSource(source graph.Source, cfg config.Config) graph.Source {
	if source.File == "" {
		return source
	}

	baseDir := cfg.BaseDir
	if baseDir == "" {
		baseDir, _ = filepath.Abs(".")
	}
	absFile, err := filepath.Abs(source.File)
	if err != nil {
		return source
	}
	relative, err := filepath.Rel(baseDir, absFile)
	if err != nil || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return source
	}
	source.File = filepath.ToSlash(relative)
	return source
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
