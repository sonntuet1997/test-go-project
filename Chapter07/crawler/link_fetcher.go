package crawler

import (
	"context"
	"io"
	"net/url"
	"strings"
	"test_project/Chapter07/pipeline"
)

type linkFetcher struct {
	urlGetter   URLGetter
	netDetector PrivateNetworkDetector
}

func (l *linkFetcher) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	if exclusionRegex.MatchString(payload.URL) {
		return nil, nil
	}
	if isPrivate, err := l.isPrivate(payload.URL); err != nil || isPrivate {
		return nil, nil
	}
	res, err := l.urlGetter.Get(payload.URL)
	if err != nil {
		return nil, nil
	}
	_, err = io.Copy(&payload.RawContent, res.Body)
	_ = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, nil
	}
	if contentType := res.Header.Get("Content-Type"); !strings.Contains(contentType, "html") {
		return nil, nil
	}
	return payload, nil
}

func (lf *linkFetcher) isPrivate(URL string) (bool, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return false, err
	}
	return lf.netDetector.IsPrivate(u.Hostname())
}

func newLinkFetcher(urlGetter URLGetter, netDetector PrivateNetworkDetector) *linkFetcher {
	return &linkFetcher{
		urlGetter:   urlGetter,
		netDetector: netDetector,
	}
}
