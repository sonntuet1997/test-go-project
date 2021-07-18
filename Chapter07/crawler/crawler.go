package crawler

import (
	"context"
	"net/http"
	"test_project/Chapter06/linkgraph/graph"
	"test_project/Chapter07/pipeline"
)

type linkSource struct {
	linkIt graph.LinkIterator
}

func (l *linkSource) Next(ctx context.Context) bool {
	return l.linkIt.Next()
}

func (l *linkSource) Payload() pipeline.Payload {
	link := l.linkIt.Link()
	p := payloadPool.Get().(*crawlerPayload)
	p.LinkID = link.ID
	p.URL = link.URL
	p.RetrievedAt = link.RetrievedAt
	return p
}

func (l *linkSource) Error() error {
	return l.linkIt.Error()
}

type nopSink struct {
}

func (n nopSink) Consume(ctx context.Context, payload pipeline.Payload) error {
	return nil
}

type URLGetter interface {
	Get(url string) (*http.Response, error)
}

type PrivateNetworkDetector interface {
	IsPrivate(host string) (bool, error)
}
