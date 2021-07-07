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
	upsertLinkQuery = `
INSERT INTO links (url, retrieved_at) VALUES ($1, $2) 
ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2) 
RETURNING id, retrieved_at`
	upsertEdgeQuery = `
INSERT INTO edges(src, dst, updated_at) VALUES ($1, $2, now())
ON CONFLICT (src, dst) DO UPDATE SET updated_at=now()
RETURNING id, updated_at
`
	findLinkQuery = `
SELECT url, retrieved_at FROM links WHERE id=$1`

	linksQuery = `
SELECT id, url, retrieved_at FROM links WHERE id >= $1 AND id < $2 AND retrieved_at < $3
`
	edgesQuery = `
SELECT id, src, dst, updated_at FROM edges WHERE src >= $1 AND src < $2 AND updated_at < $3
`
)

type CockroachDBGraph struct {
	db *sql.DB
}

func (c CockroachDBGraph) FindLink(id uuid.UUID) (*graph.Link, error) {
	row := c.db.QueryRow(findLinkQuery, id)
	link := &graph.Link{ID: id}
	if err := row.Scan(&link.URL, &link.RetrievedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.Errorf("find link: %w", graph.ErrNotFound)
		}
		return nil, xerrors.Errorf("find link: %w", err)
	}
	link.RetrievedAt = link.RetrievedAt.UTC()
	return link, nil
}

func (c CockroachDBGraph) UpsertEdge(edge *graph.Edge) error {
	row := c.db.QueryRow(upsertEdgeQuery, edge.Src, edge.Dst)
	if err := row.Scan(&edge.ID, &edge.UpdatedAt); err != nil {
		if isForeignKeyViolationError(err) {
			err = graph.ErrUnknownEdgeLinks
		}
		return xerrors.Errorf("upsert edge: %w", err)
	}
	edge.UpdatedAt = edge.UpdatedAt.UTC()
	return nil
}

func (c CockroachDBGraph) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	panic("implement me")
}

func (c CockroachDBGraph) Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	rows, err := c.db.Query(linksQuery, fromID, toID, retrievedBefore.UTC())
	if err != nil {
		return nil, xerrors.Errorf("links: %w", err)
	}
	return &linkIterator{
		rows: rows,
	}, nil
}

func (c CockroachDBGraph) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	rows, err := c.db.Query(edgesQuery, fromID, toID, updatedBefore.UTC())
	if err != nil {
		return nil, xerrors.Errorf("edges: %w", err)
	}
	return &edgeIterator{
		rows: rows,
	}, nil
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
