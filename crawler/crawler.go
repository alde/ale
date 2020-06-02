package crawler

import (
	"net/http"
	"net/url"
)

type Crawler interface {
	InitiateCrawl(buildURI string, buildID string)
	logBuildLogs(buildID string, uri *url.URL)
	updateState(buildID string)
	crawlBuild(uri *url.URL)
}

// HTTPGetter is an interface only requiring Get from http.Client
type HTTPGetter interface {
	Get(uri string) (*http.Response, error)
	Do(request *http.Request) (*http.Response, error)
}
