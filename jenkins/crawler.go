package jenkins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
)

// CrawlJenkins initiates the crawler
func CrawlJenkins(conf *config.Config, db db.Database, buildURI string, buildID string) {
	uri0 := strings.Join([]string{strings.TrimRight(buildURI, "/"), "wfapi", "describe"}, "/")
	uri, _ := url.Parse(uri0)
	processChan := make(chan string, 1)
	stateChan := make(chan *ale.JenkinsData, 1)
	go updateState(stateChan, processChan, db, buildID)
	go crawlBuild(processChan, stateChan, uri)
	processChan <- buildID
}

func updateState(stateChan <-chan *ale.JenkinsData, processChan chan<- string, db db.Database, buildID string) {
	for {
		select {
		case jdata := <-stateChan:
			logrus.Debug("got request to update the state")
			if err := db.Put(jdata, buildID); err != nil {
				logrus.WithError(err).Error("unable to add to database")
			}
			logrus.WithField("build_id", buildID).Info("database updated")

			if jdata.Status == "" || jdata.Status == "IN_PROGRESS" {
				go func() {
					logrus.Debug("sleeping for 5 seconds before requerying")
					time.Sleep(5 * time.Second)
					processChan <- buildID
				}()
			}
		}
	}
}

func crawlBuild(processChan <-chan string, stateChan chan<- *ale.JenkinsData, uri *url.URL) {
	for {
		select {
		case buildID := <-processChan:
			jd := &ale.JobData{}
			resp, err := http.Get(uri.String())
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
			jdata := extractLogs(jd, buildID, uri)
			logrus.Info("extracted jenkins data")
			stateChan <- jdata
			logrus.Debug("data sent to stateChannel")
		}
	}

}

func crawlJobStage(buildURL *url.URL, link string) ale.JobExecution {
	stageLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   link,
	}
	resp, err := http.Get(stageLink.String())
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

func crawlExecutionLogs(execution *ale.JobExecution, buildURL *url.URL) []*ale.JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   execution.Links.Log.Href,
	}
	nodeLog := extractNodeLogs(logLink)
	return []*ale.JenkinsStage{
		{
			Status:    nodeLog.NodeStatus,
			Name:      execution.Name,
			LogLength: nodeLog.Length,
			LogText:   nodeLog.Text,
			StartTime: execution.StartTimeMillis,
		},
	}
}

func extractLogsFromFlowNode(node *ale.StageFlowNode, buildURL *url.URL, ename string) *ale.JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   node.Links.Log.Href,
	}
	nodeLog := extractNodeLogs(logLink)
	return &ale.JenkinsStage{
		Status:    nodeLog.NodeStatus,
		Name:      fmt.Sprintf("%s - %s", ename, node.Name),
		LogLength: nodeLog.Length,
		LogText:   nodeLog.Text,
		StartTime: node.StartTimeMillis,
	}
}

func extractNodeLogs(logLink *url.URL) *ale.NodeLog {
	resp, err := http.Get(logLink.String())
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var nodeLog ale.NodeLog
	err = json.Unmarshal(body, &nodeLog)
	if err != nil {
		logrus.Error(err)
	}
	return &nodeLog
}

func crawlStageFlowNodesLogs(execution *ale.JobExecution, buildURL *url.URL) []*ale.JenkinsStage {
	logs := []*ale.JenkinsStage{}
	for _, node := range execution.StageFlowNodes {
		if node.Links.Log.Href == "" {
			continue
		}
		logs = append(logs, extractLogsFromFlowNode(&node, buildURL, execution.Name))
	}
	return logs
}

func extractLogsFromExecution(execution *ale.JobExecution, buildURL *url.URL) []*ale.JenkinsStage {
	if execution.Links.Log.Href == "" {
		return crawlStageFlowNodesLogs(execution, buildURL)
	}
	return crawlExecutionLogs(execution, buildURL)
}

func extractLogs(jd *ale.JobData, buildID string, buildURL *url.URL) *ale.JenkinsData {
	var stages []*ale.JenkinsStage
	for _, stage := range jd.Stages {
		execution := crawlJobStage(buildURL, stage.Links.Self.Href)
		stages = append(stages, extractLogsFromExecution(&execution, buildURL)...)
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
