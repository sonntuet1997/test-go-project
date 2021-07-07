package memory

import (
	"github.com/google/uuid"
	"golang.org/x/xerrors"
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

func (s *InMemoryGraph) UpsertLink(link *graph.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.linkURLIndex[link.URL]; existing != nil {
		link.ID = existing.ID
		origTs := existing.RetrievedAt
		*existing = *link
		if origTs.After(existing.RetrievedAt) {
			existing.RetrievedAt = origTs
		}
		return nil
	}
	for {
		link.ID = uuid.New()
		if s.links[link.ID] == nil {
			break
		}
	}
	lCopy := new(graph.Link)
	*lCopy = *link
	s.linkURLIndex[lCopy.URL] = lCopy
	s.links[lCopy.ID] = lCopy
	return nil
}

func (s *InMemoryGraph) FindLink(id uuid.UUID) (*graph.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exist := s.links[id]
	if !exist {
		return nil, xerrors.Errorf("find link: %w", graph.ErrNotFound)
	}

	lCopy := new(graph.Link)
	*lCopy = *result
	return lCopy, nil
}

func (s *InMemoryGraph) UpsertEdge(edge *graph.Edge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, srcExists := s.links[edge.Src]
	_, dstExists := s.links[edge.Dst]
	if !srcExists || !dstExists {
		return xerrors.Errorf("upsert edge: %w", graph.ErrUnknownEdgeLinks)
	}

	for _, edgeID := range s.linkEdgeMap[edge.Src] {
		existingEdge := s.edges[edgeID]
		if existingEdge.Src == edge.Src && existingEdge.Dst == edge.Dst {
			existingEdge.UpdatedAt = time.Now()
			*edge = *existingEdge
			return nil
		}
	}

	for {
		edge.ID = uuid.New()
		if s.edges[edge.ID] == nil {
			break
		}
	}

	edge.UpdatedAt = time.Now()
	eCopy := new(graph.Edge)
	*eCopy = *edge
	s.edges[eCopy.ID] = eCopy
	s.linkEdgeMap[eCopy.Src] = append(s.linkEdgeMap[eCopy.Src], eCopy.ID)
	return nil

}
func (s *InMemoryGraph) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var retain edgeList
	for _, edgeID := range s.linkEdgeMap[fromID] {
		edge := s.edges[edgeID]
		if edge.UpdatedAt.Before(updatedBefore) {
			delete(s.edges, edgeID)
		} else {
			retain = append(retain, edgeID)
		}
	}
	s.linkEdgeMap[fromID] = retain
	return nil
}

func (s *InMemoryGraph) Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	from, to := fromID.String(), toID.String()
	var list []*graph.Link
	for linkID, link := range s.links {
		if id := linkID.String(); id < to && id >= from && link.RetrievedAt.Before(retrievedBefore) {
			list = append(list, link)
		}
	}
	return &linkIterator{s: s, links: list}, nil
}

func (s *InMemoryGraph) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	from, to := fromID.String(), toID.String()
	var list []*graph.Edge
	for linkID := range s.links {
		if id := linkID.String(); id >= to || id < from {
			continue
		}
		for _, edgeID := range s.linkEdgeMap[linkID] {
			if edge := s.edges[edgeID]; edge.UpdatedAt.Before(updatedBefore) {
				list = append(list, edge)
			}
		}
	}
	return &edgeIterator{s: s, edges: list}, nil
}
