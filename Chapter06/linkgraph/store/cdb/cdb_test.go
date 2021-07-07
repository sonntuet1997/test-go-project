package cdb

import (
	"database/sql"
	gc "gopkg.in/check.v1"
	"os"
	"test_project/Chapter06/linkgraph/graph/graphtest"
	"testing"
)

var _ = gc.Suite(new(DatabaseGraphTestSuite))

func Test(t *testing.T) { gc.TestingT(t) }

type DatabaseGraphTestSuite struct {
	graphtest.SuiteBase
	db *sql.DB
}

func (s *DatabaseGraphTestSuite) SetUpTest(c *gc.C) {
	dsn := os.Getenv("CDB_DSN")
	if dsn == "" {
		c.Skip("Missing CDB_DSN envvar; skipping cockroachdb-backed graph test suite")
	}
	g, err := NewCockroachDBGraph(dsn)
	c.Assert(err, gc.IsNil)
	s.SetGraph(g)
	s.db = g.db
}
