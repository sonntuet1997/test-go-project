package memory

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"sync"
	"test_project/Chapter06/textindexer/index"
	"time"
)

const batchSize = 10

type InMemoryBleveIndexer struct {
	mu   sync.RWMutex
	docs map[string]*index.Document

	idx bleve.Index
}
type bleveDoc struct {
	Title    string
	Content  string
	PageRank float64
}

func NewInMemoryBleveIndexer() (*InMemoryBleveIndexer, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	return &InMemoryBleveIndexer{
		idx:  idx,
		docs: make(map[string]*index.Document),
	}, nil
}

func (i *InMemoryBleveIndexer) Index(doc *index.Document) error {
	if doc.LinkID == uuid.Nil {
		return xerrors.Errorf("index: %w", index.ErrMissingLinkID)
	}
	doc.IndexedAt = time.Now()
	dCopy := copyDoc(doc)
	key := dCopy.LinkID.String()
	i.mu.Lock()
	defer i.mu.Unlock()
	if savedDoc, exists := i.docs[key]; exists {
		dCopy.PageRank = savedDoc.PageRank
	}
	if err := i.idx.Index(key, makeBleveDoc(dCopy)); err != nil {
		return xerrors.Errorf("index: %w", err)
	}
	i.docs[key] = dCopy
	return nil
}

func copyDoc(doc *index.Document) *index.Document {
	dCopy := &index.Document{}
	*dCopy = *doc
	return dCopy
}

func (i *InMemoryBleveIndexer) FindByID(linkID uuid.UUID) (*index.Document, error) {
	return i.findByID(linkID.String())
}

func (i *InMemoryBleveIndexer) findByID(linkID string) (*index.Document, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if doc, found := i.docs[linkID]; found {
		return copyDoc(doc), nil
	}
	return nil, xerrors.Errorf("find by ID: %w", index.ErrNotFound)
}

func (i *InMemoryBleveIndexer) Search(q index.Query) (index.Iterator, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	var bq query.Query
	switch q.Type {
	case index.QueryTypeMatch:
		bq = bleve.NewMatchQuery(q.Expression)
	case index.QueryTypePhrase:
		bq = bleve.NewMatchPhraseQuery(q.Expression)
	}
	searchReq := bleve.NewSearchRequest(bq)
	searchReq.SortBy([]string{"-PageRank", "-_score"})
	searchReq.Size = batchSize
	searchReq.From = int(q.Offset)
	rs, err := i.idx.Search(searchReq)
	if err != nil {
		return nil, xerrors.Errorf("search: %w", err)
	}
	return &bleveIterator{idx: i, searchReq: searchReq, rs: rs, cumIdx: q.Offset}, nil
}

func (i *InMemoryBleveIndexer) UpdateScore(linkID uuid.UUID, score float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	doc, found := i.docs[linkID.String()]
	if !found {
		doc = &index.Document{LinkID: linkID}
		i.docs[linkID.String()] = doc
	}
	doc.PageRank = score
	if err := i.idx.Index(linkID.String(), makeBleveDoc(doc)); err != nil {
		return xerrors.Errorf("update score: %w", err)
	}
	return nil
}

func (i *InMemoryBleveIndexer) Close() error {
	return i.idx.Close()
}

func makeBleveDoc(d *index.Document) bleveDoc {
	return bleveDoc{
		Title:    d.Title,
		Content:  d.Content,
		PageRank: d.PageRank,
	}
}
