package crawler

import (
	"context"
	"net/http"
	"test_project/Chapter06/linkgraph/graph"
	"test_project/Chapter06/textindexer/index"
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

type Crawler struct {
	p *pipeline.Pipeline
}
type Indexer interface {
	// Index inserts a new document to the index or updates the index entry
	// for and existing document.
	Index(doc *index.Document) error
}

type Config struct {
	PrivateNetworkDetector PrivateNetworkDetector
	URLGetter              URLGetter
	Graph                  graph.Graph
	Indexer                Indexer
	FetchWorkers           int
}

func assembleCrawlerPipeline(cfg Config) *pipeline.Pipeline {
	return pipeline.New(
		pipeline.FixedWorkerPool(
			newLinkFetcher(cfg.URLGetter, cfg.PrivateNetworkDetector),
			cfg.FetchWorkers,
		),
		pipeline.FIFO(newLinkExtractor(cfg.PrivateNetworkDetector)),
		pipeline.FIFO(newTextExtractor()),
		pipeline.Broadcast(
			newGraphUpdater(cfg.Graph),
			newTextIndexer(cfg.Indexer),
		),
	)
}

func NewCrawler(cfg Config) *Crawler {
	return &Crawler{
		p: assembleCrawlerPipeline(cfg),
	}
}

func (c *Crawler) Crawl(ctx context.Context, linkIt graph.LinkIterator) (int, error) {
	sink := new(countingSink)
	err := c.p.Process(ctx, &linkSource{linkIt: linkIt}, sink)
	return sink.getCount(), err
}

type countingSink struct {
	count int
}

func (s *countingSink) Consume(_ context.Context, p pipeline.Payload) error {
	s.count++
	return nil
}

func (s *countingSink) getCount() int {
	// The broadcast split-stage sends out two payloads for each incoming link
	// so we need to divide the total count by 2.
	return s.count / 2
}
