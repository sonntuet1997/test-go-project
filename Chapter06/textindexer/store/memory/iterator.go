package memory

import (
	"github.com/blevesearch/bleve"
	"test_project/Chapter06/textindexer/index"
)

type bleveIterator struct {
	idx        *InMemoryBleveIndexer
	searchReq  *bleve.SearchRequest
	rs         *bleve.SearchResult
	rsIdx      int
	cumIdx     uint64
	latchedDoc *index.Document
	lastErr    error
}

func (l *bleveIterator) Next() bool {
	if l.lastErr != nil || l.rs == nil || l.cumIdx >= l.rs.Total {
		return false
	}
	if l.rsIdx >= l.rs.Hits.Len() {
		l.searchReq.From += l.searchReq.Size
		if l.rs, l.lastErr = l.idx.idx.Search(l.searchReq); l.lastErr != nil {
			return false
		}
		l.rsIdx = 0
	}
	nextID := l.rs.Hits[l.rsIdx].ID
	if l.latchedDoc, l.lastErr = l.idx.findByID(nextID); l.lastErr != nil {
		return false
	}
	l.rsIdx++
	l.cumIdx++
	return true
}

func (l *bleveIterator) Document() *index.Document {
	return copyDoc(l.latchedDoc)
}

func (i bleveIterator) TotalCount() uint64 {
	return i.rs.Total
}

// Error implements graph.LinkIterator.
func (i *bleveIterator) Error() error {
	return i.lastErr
}

// Close implements graph.LinkIterator.
func (i *bleveIterator) Close() error {
	return nil
}
