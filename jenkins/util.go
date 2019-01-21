package jenkins

import (
	"net/url"
	"sort"
	"strings"
)

// JobURLToAPI converts a Jenkins Job URL to it's API url
func JobURLToAPI(jobURL string) (*url.URL, error) {
	uri := strings.Join([]string{strings.TrimRight(jobURL, "/"), "wfapi", "describe"}, "/")
	return url.Parse(uri)
}

// ExtractLogs extracts logs
func ExtractLogs(jd *JobData, buildID string, buildURL *url.URL) *JenkinsData {
	var stages []*JenkinsStage
	jdata := &JenkinsData{}
	for jdata.Status == "IN_PROGRESS" || jdata.Status == "" {
		for _, stage := range jd.Stages {
			execution := crawlJobStage(buildURL, stage.Links.Self.Href)
			stages = append(stages, extractLogsFromExecution(&execution, buildURL)...)
		}
		jdata.Status = jd.Status
		jdata.Name = jd.Name
		jdata.ID = jd.ID
		jdata.BuildID = buildID
	}
	sort.Slice(stages[:], func(i, j int) bool {
		return stages[i].StartTime < stages[j].StartTime
	})
	jdata.Stages = stages
	return jdata
}
