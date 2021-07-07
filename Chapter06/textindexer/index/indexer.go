package index

import "github.com/google/uuid"

type Query struct {
	Type       QueryType
	Expression string
	Offset     uint64
}

type QueryType uint8

const (
	QueryTypeMatch QueryType = iota
	QueryTypePhrase
)

type Iterator interface {
	Close() error
	Next() bool
	Error() error
	Document() *Document
	TotalCount() uint64
}

type Indexer interface {
	Index(doc *Document) error
	FindByID(linkID uuid.UUID) (*Document, error)
	Search(query Query) (Iterator, error)
	UpdateScore(linkID uuid.UUID, score float64) error
}
