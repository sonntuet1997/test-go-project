package memory

import (
	"github.com/google/uuid"
	"sync"
	"test_project/Chapter06/linkgraph/graph"
	"time"
)

// Compile-time check for ensuring InMemoryGraph implements Graph.
var _ graph.Graph = (*InMemoryGraph)(nil)

type edgeList []uuid.UUID

type InMemoryGraph struct {
	mu sync.RWMutex

	links map[uuid.UUID]*graph.Link
	edges map[uuid.UUID]*graph.Edge

	linkURLIndex map[string]*graph.Link
	linkEdgeMap  map[uuid.UUID]edgeList
}

func NewInMemoryGraph() *InMemoryGraph {
	return &InMemoryGraph{
		links:        make(map[uuid.UUID]*graph.Link),
		edges:        make(map[uuid.UUID]*graph.Edge),
		linkURLIndex: make(map[string]*graph.Link),
		linkEdgeMap:  make(map[uuid.UUID]edgeList),
	}
}

func (receiver *InMemoryGraph) UpsertLink(link *graph.Link) error {
	existing, _ := receiver.linkURLIndex[link.URL]

	if link.ID == uuid.Nil {
		link.ID = existing.ID
		origTs := existing.RetrievedAt
		*existing = *link
		if origTs.After(existing.RetrievedAt) {
			existing.RetrievedAt = origTs
		}
		return nil
	}
	return nil
}

func (receiver *InMemoryGraph) FindLink(id uuid.UUID) (*graph.Link, error) {
	return nil, nil

}

func (receiver *InMemoryGraph) UpsertEdge(edge *graph.Edge) error {
	return nil

}
func (receiver *InMemoryGraph) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	return nil

}

func (receiver *InMemoryGraph) Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	return nil, nil

}
func (receiver *InMemoryGraph) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	return nil, nil
}
