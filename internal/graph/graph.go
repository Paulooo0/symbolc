package graph

type Graph struct {
	Symbols      map[SymbolID]Symbol
	Dependencies []Dependecy
}

func NewGraph() *Graph {
	return &Graph{
		Symbols: make(map[SymbolID]Symbol)}
}

func (g *Graph) AddSymbol(symbol Symbol) {
	g.Symbols[symbol.ID] = symbol
}
