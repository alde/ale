package jenkins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alde/ale/config"
	"github.com/kardianos/osext"
)

func CrawlJenkins(conf *config.Config, buildURI string, buildID string) {
	uri, _ := JobURLToAPI(buildURI)
	processChan := make(chan bool)
	stateChan := make(chan *JenkinsData)
	go updateState(stateChan, conf, buildID)
	go crawlBuild(processChan, stateChan, uri, buildID)
	processChan <- true
}

func updateState(stateChan <-chan *JenkinsData, conf *config.Config, buildID string) {
	for {
		select {
		case jdata := <-stateChan:
			logrus.Infof("%+v", jdata)
			b, _ := json.MarshalIndent(jdata, "", "\t")
			folder, _ := osext.ExecutableFolder()
			ioutil.WriteFile(fmt.Sprintf("%s/out_%s.json", folder, buildID), b, 0644)
			logrus.Infof("%s", string(b))
		}
	}
}

func crawlBuild(processChan chan bool, stateChan chan<- *JenkinsData, uri *url.URL, buildID string) {
	for {
		select {
		case <-processChan:
			jd := &JobData{}
			resp, err := http.Get(uri.String())
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &jd)
			logrus.WithField("uri", uri.String()).Info("crawling jenkins API")
			jdata := ExtractLogs(jd, buildID, uri)
			stateChan <- jdata
			if jdata.Status == "" || jdata.Status == "IN_PROGRESS" {
				for range time.Tick(5 * time.Second) {
					processChan <- true
				}
			}
		}
	}

}

func crawlJobStage(buildURL *url.URL, link string) JobExecution {
	stageLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   link,
	}
	logrus.WithField("uri", stageLink.String()).Info("crawling jenkins API")
	resp, err := http.Get(stageLink.String())
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var JobExecution JobExecution
	err = json.Unmarshal(body, &JobExecution)
	if err != nil {
		logrus.Error(err)
	}
	return JobExecution
}

func crawlExecutionLogs(execution *JobExecution, buildURL *url.URL) []*JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   execution.Links.Log.Href,
	}
	nodeLog := extractNodeLogs(logLink)
	return []*JenkinsStage{
		&JenkinsStage{
			Status:    nodeLog.NodeStatus,
			Name:      execution.Name,
			LogLength: nodeLog.Length,
			LogText:   nodeLog.Text,
			StartTime: execution.StartTimeMillis,
		},
	}
}

func extractLogsFromFlowNode(node *StageFlowNode, buildURL *url.URL, ename string) *JenkinsStage {
	logLink := &url.URL{
		Scheme: buildURL.Scheme,
		Host:   buildURL.Host,
		Path:   node.Links.Log.Href,
	}
	nodeLog := extractNodeLogs(logLink)
	return &JenkinsStage{
		Status:    nodeLog.NodeStatus,
		Name:      fmt.Sprintf("%s - %s", ename, node.Name),
		LogLength: nodeLog.Length,
		LogText:   nodeLog.Text,
		StartTime: node.StartTimeMillis,
	}
}

func extractNodeLogs(logLink *url.URL) *NodeLog {
	logrus.WithField("uri", logLink.String()).Info("crawling jenkins API")
	resp, err := http.Get(logLink.String())
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var nodeLog NodeLog
	err = json.Unmarshal(body, &nodeLog)
	if err != nil {
		logrus.Error(err)
	}
	return &nodeLog
}

func crawlStageFlowNodesLogs(execution *JobExecution, buildURL *url.URL) []*JenkinsStage {
	logs := []*JenkinsStage{}
	for _, node := range execution.StageFlowNodes {
		if node.Links.Log.Href == "" {
			continue
		}
		logs = append(logs, extractLogsFromFlowNode(&node, buildURL, execution.Name))
	}
	return logs
}

func extractLogsFromExecution(execution *JobExecution, buildURL *url.URL) []*JenkinsStage {
	if execution.Links.Log.Href == "" {
		return crawlStageFlowNodesLogs(execution, buildURL)
	}
	return crawlExecutionLogs(execution, buildURL)
}
