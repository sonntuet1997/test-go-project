package pipeline

import (
	"context"
	"fmt"
	gc "gopkg.in/check.v1"
	"testing"
)

var _ = gc.Suite(new(PipelineTestSuite))

type PipelineTestSuite struct{}
type testStage struct {
	c            *gc.C
	dropPayloads bool
	err          error
}

func (t testStage) Run(ctx context.Context, s StageParams) {
	defer func() {
		t.c.Logf("[stage %d] exiting", s.StageIndex())
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-s.Input():
			if !ok {
				return
			}
			t.c.Logf("[stage %d] received payload: %v", s.StageIndex(), p)
			if t.err != nil {
				t.c.Logf("[stage %d] emit error: %v", s.StageIndex(), t.err)
				s.Error() <- t.err
				return
			}
			if t.dropPayloads {
				t.c.Logf("[stage %d] dropping payload: %v", s.StageIndex(), p)
				p.MarkAsProcessed()
				continue
			}
			t.c.Logf("[stage %d] emitting payload: %v", s.StageIndex(), p)
			select {
			case <-ctx.Done():
				return
			case s.Output() <- p:
			}
		}
	}
}

func Test(t *testing.T) { gc.TestingT(t) }

func (s *PipelineTestSuite) TestDataFlow(c *gc.C) {
	stages := make([]StageRunner, 10)
	for i := 0; i < len(stages); i++ {
		stages[i] = testStage{
			c: c,
		}
	}
	src := &sourceStub{data: stringPayloads(3)}
	sink := new(sinkStub)
	p := New(stages...)
	err := p.Process(context.TODO(), src, sink)
	c.Assert(err, gc.IsNil)
	c.Assert(sink.data, gc.DeepEquals, src.data)
	assertAllProcessed(c, src.data)
}
func assertAllProcessed(c *gc.C, payloads []Payload) {
	for i, p := range payloads {
		payload := p.(*stringPayload)
		c.Assert(payload.processed, gc.Equals, true, gc.Commentf("payload %d not processed", i))
	}
}

type sourceStub struct {
	index int
	data  []Payload
	err   error
}

func (s *sourceStub) Next(context.Context) bool {
	if s.err != nil || s.index == len(s.data) {
		return false
	}

	s.index++
	return true
}
func (s *sourceStub) Error() error { return s.err }
func (s *sourceStub) Payload() Payload {
	return s.data[s.index-1]
}

type sinkStub struct {
	data []Payload
	err  error
}

func (s *sinkStub) Consume(_ context.Context, p Payload) error {
	s.data = append(s.data, p)
	return s.err
}

type stringPayload struct {
	processed bool
	val       string
}

func (s *stringPayload) Clone() Payload   { return &stringPayload{val: s.val} }
func (s *stringPayload) MarkAsProcessed() { s.processed = true }
func (s *stringPayload) String() string   { return s.val }

func stringPayloads(numValues int) []Payload {
	out := make([]Payload, numValues)
	for i := 0; i < len(out); i++ {
		out[i] = &stringPayload{val: fmt.Sprint(i)}
	}
	return out
}
