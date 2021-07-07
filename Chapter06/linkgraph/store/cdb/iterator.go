package cdb

import (
	"database/sql"
	"test_project/Chapter06/linkgraph/graph"
)

type linkIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedLink *graph.Link
}

func (l *linkIterator) Next() bool {
	if l.lastErr != nil || !l.rows.Next() {
		return false
	}
	t := new(graph.Link)
	l.lastErr = l.rows.Scan(&t.ID, &t.URL, &t.RetrievedAt)
	if l.lastErr != nil {
		return false
	}
	t.RetrievedAt = t.RetrievedAt.UTC()
	l.latchedLink = t
	return true
}

func (l *linkIterator) Link() *graph.Link {
	result := new(graph.Link)
	*result = *l.latchedLink
	return result
}

// Error implements graph.LinkIterator.
func (i *linkIterator) Error() error {
	return i.lastErr
}

// Close implements graph.LinkIterator.
func (i *linkIterator) Close() error {
	return nil
}

type edgeIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedEdge *graph.Edge
}

func (e *edgeIterator) Next() bool {
	if e.lastErr != nil || !e.rows.Next() {
		return false
	}
	r := new(graph.Edge)
	if e.lastErr = e.rows.Scan(&r.ID, &r.Src, &r.Dst, &r.UpdatedAt); e.lastErr != nil {
		return false
	}
	r.UpdatedAt = r.UpdatedAt.UTC()
	e.latchedEdge = r
	return true
}

func (e *edgeIterator) Edge() *graph.Edge {
	eCopy := new(graph.Edge)
	*eCopy = *e.latchedEdge
	return eCopy
}

// Error implements graph.LinkIterator.
func (e *edgeIterator) Error() error {
	return e.lastErr
}

// Close implements graph.LinkIterator.
func (e *edgeIterator) Close() error {
	return nil
}
