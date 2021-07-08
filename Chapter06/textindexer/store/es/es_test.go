package es

import (
	gc "gopkg.in/check.v1"
	"os"
	"strings"
	"test_project/Chapter06/textindexer/index/indextest"
	"testing"
)

var _ = gc.Suite(new(ElasticsearchTestSuite))

type ElasticsearchTestSuite struct {
	indextest.SuiteBase
	idx *ElasticSearchIndexer
}

func Test(t *testing.T) {
	gc.TestingT(t)
}

func (s *ElasticsearchTestSuite) SetUpSuite(c *gc.C) {
	nodeList := os.Getenv("ES_NODES")
	if nodeList == "" {
		c.Skip("Missing ES_NODES envvar; skipping elasticsearch-backed index test suite")
	}
	idx, err := NewElasticSearchIndexer(strings.Split(nodeList, ","), true)
	c.Assert(err, gc.IsNil)
	s.SetIndexer(idx)
	s.idx = idx
}

func (s *ElasticsearchTestSuite) SetUpTest(c *gc.C) {
	if s.idx.es != nil {
		_, err := s.idx.es.Indices.Delete([]string{indexName})
		c.Assert(err, gc.IsNil)
		err = ensureIndex(s.idx.es)
		c.Assert(err, gc.IsNil)
	}
}
