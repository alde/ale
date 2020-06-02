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

// HTTP is an interface only requiring Get and Do from http.Client
type HTTP interface {
	Get(uri string) (*http.Response, error)
	Do(request *http.Request) (*http.Response, error)
}
