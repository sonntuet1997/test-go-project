package memory

import (
	gc "gopkg.in/check.v1"
	"test_project/Chapter06/linkgraph/graph/graphtest"
	"testing"
)

var x = gc.Suite(new(InMemoryGraphTestSuite))

func Test(t *testing.T) { gc.TestingT(t) }

type InMemoryGraphTestSuite struct {
	graphtest.SuiteBase
}

func (s *InMemoryGraphTestSuite) SetUpTest(c *gc.C) {
	s.SetGraph(NewInMemoryGraph())
}
