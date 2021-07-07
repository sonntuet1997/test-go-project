package cdb

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/xerrors"
	"test_project/Chapter06/linkgraph/graph"
	"time"
)

var (
	upsertLinkQuery = "INSERT INTO links (url, retrieved_at) VALUES ($1, $2) ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, 2)"
)

type CockroachDBGraph struct {
	db *sql.DB
}

func (c CockroachDBGraph) FindLink(id uuid.UUID) (*graph.Link, error) {
	panic("implement me")
}

func (c CockroachDBGraph) UpsertEdge(edge *graph.Edge) error {
	panic("implement me")
}

func (c CockroachDBGraph) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	panic("implement me")
}

func (c CockroachDBGraph) Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	panic("implement me")
}

func (c CockroachDBGraph) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	panic("implement me")
}

func (c CockroachDBGraph) UpsertLink(link *graph.Link) error {
	row := c.db.QueryRow(upsertLinkQuery, link.URL, link.RetrievedAt.UTC())
	if err := row.Scan(&link.ID, &link.RetrievedAt); err != nil {
		return xerrors.Errorf("upsert link: %w", err)
	}
	link.RetrievedAt = link.RetrievedAt.UTC()
	return nil
}

func NewCockroachDBGraph(dsn string) (*CockroachDBGraph, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &CockroachDBGraph{db: db}, nil
}

// isForeignKeyViolationError returns true if err indicates a foreign key
// constraint violation.
func isForeignKeyViolationError(err error) bool {
	pqErr, valid := err.(*pq.Error)
	if !valid {
		return false
	}

	return pqErr.Code.Name() == "foreign_key_violation"
}
