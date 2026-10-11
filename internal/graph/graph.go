package graph

type Graph struct {
	Symbols      map[SymbolID]*Symbol
	Dependencies []*Dependecy
}

func NewGraph() *Graph {
	return &Graph{
		Symbols: make(map[SymbolID]*Symbol)}
}

func (g *Graph) AddSymbol(symbol *Symbol) {
	g.Symbols[symbol.ID] = symbol
}

func (g *Graph) AddDependency(dependecy *Dependecy) {
	g.Dependencies = append(g.Dependencies, dependecy)
}

func (g *Graph) AddMethod(receiverID, methodID SymbolID) {
	receiver, ok := g.Symbols[receiverID]
	if !ok {
		return
	}

	for _, id := range receiver.Methods {
		if id == methodID {
			return
		}
	}

	receiver.Methods = append(receiver.Methods, methodID)
}
