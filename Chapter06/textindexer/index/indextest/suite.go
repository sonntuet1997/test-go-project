package indextest

import (
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	gc "gopkg.in/check.v1"
	"test_project/Chapter06/textindexer/index"
	"time"
)

type SuiteBase struct {
	idx index.Indexer
}

func (s *SuiteBase) SetIndexer(i index.Indexer) {
	s.idx = i
}

func (s *SuiteBase) TestIndex(c *gc.C) {
	doc := &index.Document{
		LinkID:    uuid.New(),
		URL:       "http://example.com",
		Title:     "Illustrious examples",
		Content:   "Lorem ipsum dolor",
		IndexedAt: time.Now().Add(-12 * time.Hour).UTC(),
	}
	err := s.idx.Index(doc)
	c.Assert(err, gc.IsNil, gc.Commentf("TestIndex fail error"))

	updatedDoc := &index.Document{
		LinkID:    doc.LinkID,
		URL:       "http://example.com",
		Title:     "A more exciting title",
		Content:   "Ovidius poeta in terra pontica",
		IndexedAt: time.Now().UTC(),
	}
	err = s.idx.Index(updatedDoc)
	c.Assert(err, gc.IsNil, gc.Commentf("Update fail error"))

	incompleteDoc := &index.Document{
		URL: "http://example.com",
	}

	err = s.idx.Index(incompleteDoc)
	c.Assert(xerrors.Is(err, index.ErrMissingLinkID), gc.Equals, true)
}

func (s SuiteBase) TestFindByID(c *gc.C) {
	doc := &index.Document{
		LinkID:    uuid.New(),
		URL:       "http://example.com",
		Title:     "Illustrious examples",
		Content:   "Lorem ipsum dolor",
		IndexedAt: time.Now().Add(-12 * time.Hour).UTC(),
	}
	err := s.idx.Index(doc)
	c.Assert(err, gc.IsNil, gc.Commentf("TestIndex fail error"))

	found, e := s.idx.FindByID(doc.LinkID)
	c.Assert(e, gc.IsNil, gc.Commentf("Error in TestFindByID"))
	c.Assert(found, gc.DeepEquals, doc, gc.Commentf("Not original"))

	_, err = s.idx.FindByID(uuid.New())
	c.Assert(xerrors.Is(err, index.ErrNotFound), gc.Equals, true, gc.Commentf("sdweqwxzdas"))
}
