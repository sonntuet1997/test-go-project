package cdb

import (
	"database/sql"
	gc "gopkg.in/check.v1"
	"os"
	"test_project/Chapter06/linkgraph/graph/graphtest"
	"testing"
)

var _ = gc.Suite(new(CockroachDBGraphTestSuite))

func Test(t *testing.T) { gc.TestingT(t) }

type CockroachDBGraphTestSuite struct {
	graphtest.SuiteBase
	db *sql.DB
}

func (s *CockroachDBGraphTestSuite) SetUpSuite(c *gc.C) {
	dsn := os.Getenv("CDB_DSN")
	if dsn == "" {
		c.Skip("Missing CDB_DSN envvar; skipping cockroachdb-backed graph test suite")
	}
	g, err := NewCockroachDBGraph(dsn)
	c.Assert(err, gc.IsNil)
	s.SetGraph(g)
	s.db = g.db
}

func (s CockroachDBGraphTestSuite) SetUpTest(c *gc.C) {
	s.flushDB(c)
}

func (s *CockroachDBGraphTestSuite) flushDB(c *gc.C) {
	_, err := s.db.Exec("DELETE FROM  links")
	c.Assert(err, gc.IsNil)
	_, err = s.db.Exec("DELETE FROM  edges")
	c.Assert(err, gc.IsNil)
}
