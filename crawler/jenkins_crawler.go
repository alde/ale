package crawler

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

	"github.com/sirupsen/logrus"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
)

// Crawler struct holds various attributes needed by the crawler
type JenkinsCrawler struct {
	database       db.Database
	config         *config.Config
	processChannel chan string
	stateChannel   chan *ale.JenkinsData
	logChannel     chan []*ale.Log
	httpClient     HTTPGetter
	r              *regexp.Regexp
	log            *logrus.Logger
}

// NewJenkinsCrawler instatiates a new crawler
func NewJenkinsCrawler(db db.Database, conf *config.Config) *JenkinsCrawler {
	r, err := regexp.Compile(conf.Crawler.JenkinsLogPattern)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create log matcher")
	}
	return &JenkinsCrawler{
		database:       db,
		config:         conf,
		processChannel: make(chan string, 1),
		stateChannel:   make(chan *ale.JenkinsData, 1),
		logChannel:     make(chan []*ale.Log, 1),
		httpClient:     http.DefaultClient,
		r:              r,
		log:            logrus.New(),
	}
}

// InitiateCrawl initiates the crawler
func (c *JenkinsCrawler) InitiateCrawl(buildURI string, buildID string) {
	uri0 := strings.Join([]string{strings.TrimRight(buildURI, "/"), "wfapi", "describe"}, "/")
	uri, _ := url.Parse(uri0)

	go c.updateState(buildID)
	go c.crawlBuild(uri)
	go c.logBuildLogs(buildID, uri)

	c.processChannel <- buildID
}

func (c *JenkinsCrawler) logBuildLogs(buildID string, uri *url.URL) {
	for {
		select {
		case jlogs := <-c.logChannel:
			logrus.Debug("got request to log the jenkins build logs")
			for _, jlog := range jlogs {
				c.printBuildLog(jlog, uri, buildID)
			}
		}
	}
}

func (c *JenkinsCrawler) printBuildLog(jlog *ale.Log, uri *url.URL, buildID string) {
	logrus.WithFields(logrus.Fields{
		"uri":             uri.String(),
		"build_id":        buildID,
		"build_timestamp": jlog.TimeStamp,
	}).Info(jlog.Line)
}

func (c *JenkinsCrawler) extractBuildLogs(jdata *ale.JenkinsData) []*ale.Log {
	var jlogs []*ale.Log
	for _, stage := range jdata.Stages {
		if stage.SubStages != nil && len(stage.SubStages) > 0 {
			for _, substage := range stage.SubStages {
				jlogs = append(jlogs, substage.Logs...)
			}
			continue
		}
		jlogs = append(jlogs, stage.Logs...)
	}
	return jlogs
}

func (c *JenkinsCrawler) updateState(buildID string) {
	for {
		select {
		case jdata := <-c.stateChannel:
			logrus.Debug("got request to update the state")
			if err := c.database.Put(jdata, buildID); err != nil {
				logrus.WithField("build_id", buildID).WithError(err).Error("unable to add to database")
			}
			logrus.WithField("build_id", buildID).Info("database updated")

			if jdata.Status == "" || jdata.Status == "IN_PROGRESS" {
				go func() {
					logrus.Debug("sleeping for 5 seconds before requerying")
					time.Sleep(5 * time.Second)
					c.processChannel <- buildID
				}()
				continue
			}

			jlogs := c.extractBuildLogs(jdata)
			logrus.Info("extracted jenkins build logs")
			c.logChannel <- jlogs
			logrus.Debug("build logs sent to logChannel")

			logrus.WithFields(logrus.Fields{
				"build_id": buildID,
				"status":   jdata.Status,
			}).Info("build finished")
			return
		}
	}
}

func (c *JenkinsCrawler) crawlBuild(uri *url.URL) {
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

func (c *JenkinsCrawler) crawlJobStage(buildURL *url.URL, link string) ale.JobExecution {
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
	if err != nil {
		logrus.Error(err)
	}
	var JobExecution ale.JobExecution
	err = json.Unmarshal(body, &JobExecution)
	if err != nil {
		logrus.Error(err)
	}
	return JobExecution
}

func (c *JenkinsCrawler) crawlExecutionLogs(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
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
		Duration:  execution.DurationMillis,
	}
}

func (c *JenkinsCrawler) extractLogsFromFlowNode(node *ale.StageFlowNode, buildURL *url.URL, ename string, flowNodesByID map[string]*ale.StageFlowNode) *ale.JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   node.Links.Log.Href,
	}
	nodeLog := c.extractNodeLogs(logLink)
	task := c.findTask(node, flowNodesByID)
	return &ale.JenkinsStage{
		Status:      nodeLog.NodeStatus,
		Name:        fmt.Sprintf("%s - %s", ename, node.Name),
		LogLength:   nodeLog.Length,
		Logs:        c.splitLogs(nodeLog.Text),
		StartTime:   node.StartTimeMillis,
		Duration:    node.DurationMillis,
		Task:        task,
		Description: node.ParameterDescription,
	}
}

func (c *JenkinsCrawler) findTask(node *ale.StageFlowNode, flowNodesByID map[string]*ale.StageFlowNode) string {
	if strings.Contains(node.ParameterDescription, "from task") {
		return strings.TrimSpace(strings.Split(node.ParameterDescription, "from task")[1])
	}
	if len(node.Parents) == 0 {
		return ""
	}
	var firstParent = flowNodesByID[node.Parents[0]]
	if firstParent == nil {
		return ""
	}
	return c.findTask(firstParent, flowNodesByID)
}

func (c *JenkinsCrawler) extractNodeLogs(logLink *url.URL) *ale.NodeLog {
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

func (c *JenkinsCrawler) crawlStageFlowNodesLogs(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
	logs := []*ale.JenkinsStage{}
	var flowNodesByID = make(map[string]*ale.StageFlowNode)
	for i := range execution.StageFlowNodes {
		node := execution.StageFlowNodes[i]
		flowNodesByID[node.ID] = &node
	}

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
		logs = append(logs, c.extractLogsFromFlowNode(&node, logLink, execution.Name, flowNodesByID))
	}
	return &ale.JenkinsStage{
		Status:    execution.Status,
		Name:      execution.Name,
		SubStages: logs,
		StartTime: execution.StartTimeMillis,
		Duration:  execution.DurationMillis,
	}
}

func (c *JenkinsCrawler) extractLogsFromExecution(execution *ale.JobExecution, buildURL *url.URL) *ale.JenkinsStage {
	logrus.WithField("id", execution.ID).Debug("crawling execution")
	if execution.StageFlowNodes != nil && len(execution.StageFlowNodes) > 0 {
		return c.crawlStageFlowNodesLogs(execution, buildURL)
	}
	return c.crawlExecutionLogs(execution, buildURL)
}

func (c *JenkinsCrawler) extractLogs(jd *ale.JobData, buildID string, buildURL *url.URL) *ale.JenkinsData {
	var stages []*ale.JenkinsStage
	for _, stage := range jd.Stages {
		execution := c.crawlJobStage(buildURL, stage.Links.Self.Href)
		stages = append(stages, c.extractLogsFromExecution(&execution, buildURL))
	}

	sort.Slice(stages[:], func(i, j int) bool {
		return stages[i].StartTime < stages[j].StartTime
	})

	return &ale.JenkinsData{
		Status:        jd.Status,
		Name:          jd.Name,
		ID:            jd.ID,
		BuildID:       buildID,
		Stages:        stages,
		Duration:      jd.DurationMillis,
		StartTime:     jd.StartTimeMillis,
		EndTime:       jd.EndTimeMillis,
		QueueDuration: jd.QueueDurationMillis,
		PauseDuration: jd.PauseDurationMillis,
	}
}

func (c *JenkinsCrawler) splitLogs(log string) []*ale.Log {
	var l []*ale.Log
	for _, part := range strings.Split(log, "\n") {
		if part == "" {
			continue
		}
		l = append(l, c.extractTimestamp(part))
	}
	return l
}

func (c *JenkinsCrawler) extractTimestamp(line string) *ale.Log {
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
