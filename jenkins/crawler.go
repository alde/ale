package jenkins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
)

// Crawler struct holds various attributes needed by the crawler
type Crawler struct {
	database       db.Database
	config         *config.Config
	processChannel chan string
	stateChannel   chan *ale.JenkinsData
	httpClient     HTTPGetter
	r              *regexp.Regexp
}

// HTTPGetter is an interface only requiring Get from http.Client
type HTTPGetter interface {
	Get(uri string) (*http.Response, error)
}

// NewCrawler instatiates a new crawler
func NewCrawler(db db.Database, conf *config.Config) *Crawler {
	r, err := regexp.Compile(conf.LogPattern)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create log matcher")
	}
	return &Crawler{
		database:       db,
		config:         conf,
		processChannel: make(chan string, 1),
		stateChannel:   make(chan *ale.JenkinsData, 1),
		httpClient:     http.DefaultClient,
		r:              r,
	}
}

// CrawlJenkins initiates the crawler
func (c *Crawler) CrawlJenkins(buildURI string, buildID string) {
	uri0 := strings.Join([]string{strings.TrimRight(buildURI, "/"), "wfapi", "describe"}, "/")
	uri, _ := url.Parse(uri0)

	go c.updateState(buildID)
	go c.crawlBuild(uri)
	c.processChannel <- buildID
}

func (c *Crawler) updateState(buildID string) {
	for {
		select {
		case jdata := <-c.stateChannel:
			logrus.Debug("got request to update the state")
			if err := c.database.Put(jdata, buildID); err != nil {
				logrus.WithError(err).Error("unable to add to database")
			}
			logrus.WithField("build_id", buildID).Info("database updated")

			if jdata.Status == "" || jdata.Status == "IN_PROGRESS" {
				go func() {
					logrus.Debug("sleeping for 5 seconds before requerying")
					time.Sleep(5 * time.Second)
					c.processChannel <- buildID
				}()
			} else {
				logrus.WithFields(logrus.Fields{
					"build_id": buildID,
					"status":   jdata.Status,
				}).Info("build finished")
				return
			}
		}
	}
}

func (c *Crawler) crawlBuild(uri *url.URL) {
	for {
		select {
		case buildID := <-c.processChannel:
			jd := &ale.JobData{}
			resp, err := c.httpClient.Get(uri.String())
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &jd)
			logrus.WithFields(logrus.Fields{
				"uri":      uri.String(),
				"build_id": buildID,
			}).Info("crawling jenkins API")
			jdata := c.extractLogs(jd, buildID, uri)
			logrus.Info("extracted jenkins data")
			c.stateChannel <- jdata
			logrus.Debug("data sent to stateChannel")
		}
	}
}

func (c *Crawler) crawlJobStage(buildURL *url.URL, link string) ale.JobExecution {
	stageLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   link,
	}
	resp, err := c.httpClient.Get(stageLink.String())
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var JobExecution ale.JobExecution
	err = json.Unmarshal(body, &JobExecution)
	if err != nil {
		logrus.Error(err)
	}
	return JobExecution
}

func (c *Crawler) crawlExecutionLogs(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   execution.Links.Log.Href,
	}
	nodeLog := c.extractNodeLogs(logLink)
	return &ale.JenkinsStage{
		Status:    nodeLog.NodeStatus,
		Name:      execution.Name,
		LogLength: nodeLog.Length,
		Logs:      c.splitLogs(nodeLog.Text),
		StartTime: execution.StartTimeMillis,
	}
}

func (c *Crawler) extractLogsFromFlowNode(node *ale.StageFlowNode, buildURL *url.URL, ename string) *ale.JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   node.Links.Log.Href,
	}
	nodeLog := c.extractNodeLogs(logLink)
	return &ale.JenkinsStage{
		Status:    nodeLog.NodeStatus,
		Name:      fmt.Sprintf("%s - %s", ename, node.Name),
		LogLength: nodeLog.Length,
		Logs:      c.splitLogs(nodeLog.Text),
		StartTime: node.StartTimeMillis,
	}
}

func (c *Crawler) extractNodeLogs(logLink *url.URL) *ale.NodeLog {
	resp, err := http.Get(logLink.String())
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var nodeLog ale.NodeLog
	err = json.Unmarshal(body, &nodeLog)
	if err != nil {
		logrus.WithError(err).WithField("url", logLink.String()).Error("unable to extract logs from node")
	}
	return &nodeLog
}

func (c *Crawler) crawlStageFlowNodesLogs(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
	logs := []*ale.JenkinsStage{}
	for _, node := range execution.StageFlowNodes {
		if node.Links.Log.Href == "" {
			continue
		}
		logLink := &url.URL{
			Scheme: buildURL.Scheme,
			Host:   buildURL.Host,
			Path:   node.Links.Log.Href,
		}
		logrus.WithFields(logrus.Fields{
			"url":  logLink,
			"node": node.ID,
		}).Debug("crawling jenkins")
		logs = append(logs, c.extractLogsFromFlowNode(&node, logLink, execution.Name))
	}
	logrus.Debugf("%+v", logs)
	return &ale.JenkinsStage{
		Status:    execution.Status,
		Name:      execution.Name,
		SubStages: logs,
		StartTime: execution.StartTimeMillis,
	}
}

func (c *Crawler) extractLogsFromExecution(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
	logrus.WithField("id", execution.ID).Debug("crowling execution")
	if execution.StageFlowNodes != nil && len(execution.StageFlowNodes) > 0 {
		return c.crawlStageFlowNodesLogs(execution, buildURL)
	}
	return c.crawlExecutionLogs(execution, buildURL)
}

func (c *Crawler) extractLogs(jd *ale.JobData, buildID string, buildURL *url.URL) *ale.JenkinsData {
	var stages []*ale.JenkinsStage
	for _, stage := range jd.Stages {
		execution := c.crawlJobStage(buildURL, stage.Links.Self.Href)
		stages = append(stages, c.extractLogsFromExecution(&execution, buildURL))
	}

	sort.Slice(stages[:], func(i, j int) bool {
		return stages[i].StartTime < stages[j].StartTime
	})

	return &ale.JenkinsData{
		Status:  jd.Status,
		Name:    jd.Name,
		ID:      jd.ID,
		BuildID: buildID,
		Stages:  stages,
	}
}

func (c *Crawler) splitLogs(log string) []*ale.Log {
	var l []*ale.Log
	for _, part := range strings.Split(log, "\n") {
		if part == "" {
			continue
		}
		l = append(l, c.extractTimestamp(part))
	}
	return l
}

func (c *Crawler) extractTimestamp(line string) *ale.Log {
	re := c.r.FindStringSubmatch(line)
	if len(re) <= 1 {
		return &ale.Log{
			Line: line,
		}
	}
	return &ale.Log{
		TimeStamp: re[1],
		Line:      re[2],
	}
}
