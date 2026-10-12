package output

import (
	"testing"

	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/graph"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMaterialize(t *testing.T) {
	g := graph.NewGraph()
	mainID := graph.SymbolID("example.com/app/cmd")
	packageID := graph.SymbolID("example.com/app")
	userID := graph.SymbolID("example.com/app.User")
	methodID := graph.SymbolID("example.com/app.User.String")

	g.AddPackage(&graph.PackageInfo{ID: mainID, Name: "main", Imports: []graph.SymbolID{packageID}, Main: true})
	g.AddPackage(&graph.PackageInfo{ID: packageID, Name: "app"})
	for _, symbol := range []*graph.Symbol{
		{ID: mainID, Name: "main", Kind: graph.KindPackage, Package: graph.Package{Name: "main", ImportPath: string(mainID)}, Source: graph.Source{File: "/project/cmd/app/main.go"}},
		{ID: packageID, Name: "app", Kind: graph.KindPackage, Package: graph.Package{Name: "app", ImportPath: string(packageID)}, Source: graph.Source{File: "/project/internal/app/app.go"}},
		{ID: userID, Name: "User", Kind: graph.KindStruct, Package: graph.Package{Name: "app", ImportPath: string(packageID)}, Source: graph.Source{File: "/project/internal/app/user.go"}},
		{ID: methodID, Name: "String", Kind: graph.KindMethod, Package: graph.Package{Name: "app", ImportPath: string(packageID)}, Source: graph.Source{File: "/project/internal/app/user.go"}},
	} {
		g.AddSymbol(symbol)
	}
	g.AddMethod(userID, methodID)

	files, err := Materialize(g, config.Config{BaseDir: "/project"})
	require.NoError(t, err)
	require.Contains(t, files, "cmd/app/index.yml")
	require.Contains(t, files, "index.yml")
	require.Contains(t, files, "cmd/app/flux-entrypoints/app.yml")
	require.Contains(t, files, "cmd/app/contracts/app/User.yml")

	var index artifactIndex
	var atlas artifactIndex
	require.NoError(t, yaml.Unmarshal(files["index.yml"], &atlas))
	require.Equal(t, []artifactRef{{ID: string(mainID), Path: "cmd/app/index.yml"}}, atlas.Executables)
	require.NoError(t, yaml.Unmarshal(files["cmd/app/index.yml"], &index))
	require.Equal(t, []artifactRef{{ID: string(packageID), Path: "flux-entrypoints/app.yml"}}, index.FluxEntrypoints)
	require.Len(t, index.Contracts, 4)

	var flux fluxEntrypoint
	require.NoError(t, yaml.Unmarshal(files["cmd/app/flux-entrypoints/app.yml"], &flux))
	require.Equal(t, packageID, flux.ID)
	require.Len(t, flux.Packages, 1)
	require.Equal(t, packageID, flux.Packages[0].ID)

	var contract contract
	require.NoError(t, yaml.Unmarshal(files["cmd/app/contracts/app/User.yml"], &contract))
	require.Equal(t, userID, contract.ID)
	require.Equal(t, []graph.SymbolID{methodID}, contract.Methods)
}

func TestMaterializeFiltersContracts(t *testing.T) {
	g := graph.NewGraph()
	g.AddSymbol(&graph.Symbol{
		ID:      "example.com/app.User",
		Name:    "User",
		Kind:    graph.KindStruct,
		Package: graph.Package{Name: "app", ImportPath: "example.com/app"},
		Source:  graph.Source{File: "/project/internal/user.go"},
	})
	g.AddSymbol(&graph.Symbol{
		ID:      "example.com/app.Hidden",
		Name:    "Hidden",
		Kind:    graph.KindStruct,
		Package: graph.Package{Name: "app", ImportPath: "example.com/app"},
		Source:  graph.Source{File: "/project/internal/generated/hidden.go"},
	})

	files, err := Materialize(g, config.Config{
		BaseDir: "/project",
		Contracts: config.ContractsConfig{
			Include: []string{"internal/**"},
			Exclude: []string{"**/generated/**"},
		},
	})
	require.NoError(t, err)
	require.Contains(t, files, "contracts/app/User.yml")
	require.NotContains(t, files, "contracts/app/Hidden.yml")
}

func TestMaterializeSeparatesExecutables(t *testing.T) {
	g := graph.NewGraph()
	mainOne := graph.SymbolID("example.com/one/cmd")
	mainTwo := graph.SymbolID("example.com/two/cmd")
	sharedOne := graph.SymbolID("example.com/one/shared")
	sharedTwo := graph.SymbolID("example.com/two/shared")

	g.AddPackage(&graph.PackageInfo{ID: mainOne, Name: "main", Imports: []graph.SymbolID{sharedOne}, Main: true})
	g.AddPackage(&graph.PackageInfo{ID: mainTwo, Name: "main", Imports: []graph.SymbolID{sharedTwo}, Main: true})
	g.AddPackage(&graph.PackageInfo{ID: sharedOne, Name: "shared"})
	g.AddPackage(&graph.PackageInfo{ID: sharedTwo, Name: "shared"})
	for _, symbol := range []*graph.Symbol{
		{ID: mainOne, Name: "main", Kind: graph.KindPackage, Package: graph.Package{Name: "main", ImportPath: string(mainOne)}, Source: graph.Source{File: "/project/cmd/one/main.go"}},
		{ID: mainTwo, Name: "main", Kind: graph.KindPackage, Package: graph.Package{Name: "main", ImportPath: string(mainTwo)}, Source: graph.Source{File: "/project/cmd/two/main.go"}},
		{ID: sharedOne, Name: "shared", Kind: graph.KindPackage, Package: graph.Package{Name: "shared", ImportPath: string(sharedOne)}, Source: graph.Source{File: "/project/one/shared/shared.go"}},
		{ID: sharedTwo, Name: "shared", Kind: graph.KindPackage, Package: graph.Package{Name: "shared", ImportPath: string(sharedTwo)}, Source: graph.Source{File: "/project/two/shared/shared.go"}},
	} {
		g.AddSymbol(symbol)
	}

	files, err := Materialize(g, config.Config{BaseDir: "/project"})
	require.NoError(t, err)
	require.Contains(t, files, "index.yml")
	require.Contains(t, files, "cmd/one/index.yml")
	require.Contains(t, files, "cmd/two/index.yml")
	require.Contains(t, files, "cmd/one/flux-entrypoints/shared.yml")
	require.Contains(t, files, "cmd/two/flux-entrypoints/shared.yml")
	require.Contains(t, files, "cmd/one/contracts/shared/package.yml")
	require.Contains(t, files, "cmd/two/contracts/shared/package.yml")

	var atlas artifactIndex
	require.NoError(t, yaml.Unmarshal(files["index.yml"], &atlas))
	require.Equal(t, []artifactRef{
		{ID: string(mainOne), Path: "cmd/one/index.yml"},
		{ID: string(mainTwo), Path: "cmd/two/index.yml"},
	}, atlas.Executables)

	again, err := Materialize(g, config.Config{BaseDir: "/project"})
	require.NoError(t, err)
	for path, data := range files {
		require.Equal(t, data, again[path], path)
	}
}
