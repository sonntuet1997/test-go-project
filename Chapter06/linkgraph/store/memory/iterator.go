package memory

import "test_project/Chapter06/linkgraph/graph"

type linkIterator struct {
	s        *InMemoryGraph
	links    []*graph.Link
	curIndex int
}

func (l *linkIterator) Next() bool {
	if l.curIndex >= len(l.links) {
		return false
	}
	l.curIndex++
	return true
}

func (l *linkIterator) Link() *graph.Link {
	l.s.mu.RLock()
	defer l.s.mu.RUnlock()

	link := l.links[l.curIndex-1]
	lCopy := new(graph.Link)
	*lCopy = *link
	return lCopy
}

// Error implements graph.LinkIterator.
func (i *linkIterator) Error() error {
	return nil
}

// Close implements graph.LinkIterator.
func (i *linkIterator) Close() error {
	return nil
}

type edgeIterator struct {
	s        *InMemoryGraph
	edges    []*graph.Edge
	curIndex int
}

func (e *edgeIterator) Next() bool {
	if e.curIndex >= len(e.edges) {
		return false
	}
	e.curIndex++
	return true
}

func (e *edgeIterator) Edge() *graph.Edge {
	e.s.mu.RLock()
	defer e.s.mu.RUnlock()

	edge := e.edges[e.curIndex-1]
	eCopy := new(graph.Edge)
	*eCopy = *edge
	return eCopy
}

// Error implements graph.LinkIterator.
func (e *edgeIterator) Error() error {
	return nil
}

// Close implements graph.LinkIterator.
func (e *edgeIterator) Close() error {
	return nil
}
