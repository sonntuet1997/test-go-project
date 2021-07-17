package crawler

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io"
	"sync"
	"test_project/Chapter07/pipeline"
	"time"
)

type crawlerPayload struct {
	LinkID      uuid.UUID
	URL         string
	RetrievedAt time.Time

	RawContent    bytes.Buffer
	NoFollowLinks []string
	Links         []string
	Title         string
	TextContent   string
}

func (c *crawlerPayload) Clone() pipeline.Payload {
	newP := payloadPool.Get().(*crawlerPayload)
	newP.LinkID = c.LinkID
	newP.URL = c.URL
	newP.RetrievedAt = c.RetrievedAt
	newP.NoFollowLinks = append([]string(nil), c.NoFollowLinks...)
	newP.Links = append([]string(nil), c.Links...)
	newP.Title = c.Title
	newP.TextContent = c.TextContent
	_, err := io.Copy(&newP.RawContent, &c.RawContent)
	if err != nil {
		panic(fmt.Sprintf("[BUG] error cloning payload raw content: %v", err))
	}
	return newP
}

func (c *crawlerPayload) MarkAsProcessed() {
	c.URL = c.URL[:0]
	c.RawContent.Reset()
	c.NoFollowLinks = c.NoFollowLinks[:0]
	c.Links = c.Links[:0]
	c.Title = c.Title[:0]
	c.TextContent = c.TextContent[:0]
	payloadPool.Put(c)
}

var (
	_           pipeline.Payload = (*crawlerPayload)(nil)
	payloadPool                  = sync.Pool{
		New: func() interface{} { return new(crawlerPayload) },
	}
)
