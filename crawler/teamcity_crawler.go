package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
	"github.com/sirupsen/logrus"
)

// Crawler struct holds various attributes needed by the crawler
type TeamCityCrawler struct {
	database       db.Database
	config         *config.Config
	processChannel chan string
	stateChannel   chan *ale.TeamCityData
	logChannel     chan []*ale.Log
	httpClient     HTTP
	r              *regexp.Regexp
	log            *logrus.Logger
	token          string
}

// NewCrawler instantiates a new crawler
func NewTeamCityCrawler(db db.Database, conf *config.Config) *TeamCityCrawler {
	r, err := regexp.Compile(conf.Crawler.TeamcityLogPattern)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create log matcher")
	}
	return &TeamCityCrawler{
		database:       db,
		config:         conf,
		processChannel: make(chan string, 1),
		stateChannel:   make(chan *ale.TeamCityData, 1),
		logChannel:     make(chan []*ale.Log, 1),
		httpClient:     http.DefaultClient,
		r:              r,
		log:            logrus.New(),
		token:          conf.Token.TCAccessToken,
	}
}

// InitiateCrawl initiates the crawler
func (c *TeamCityCrawler) InitiateCrawl(buildURI string, buildID string) {
	uri, _ := url.Parse(buildURI)

	go c.updateState(buildID)
	go c.crawlBuild(uri)
	go c.logBuildLogs(buildID, uri)

	c.processChannel <- buildID
}

func (c *TeamCityCrawler) logBuildLogs(buildID string, uri *url.URL) {
	for {
		select {
		case tcLogs := <-c.logChannel:
			logrus.Debug("Got request to log teamcity build logs.")
			for _, tclog := range tcLogs {
				c.printBuildLog(tclog, buildID)
			}
		}
	}
}

func (c *TeamCityCrawler) printBuildLog(tclog *ale.Log, buildID string) {
	logrus.WithFields(logrus.Fields{
		"build_id":        buildID,
		"build_timestamp": tclog.TimeStamp,
	}).Info(tclog.Line)
}

func (c *TeamCityCrawler) updateState(buildID string) {
	for {
		select {
		case tcdata := <-c.stateChannel:
			logrus.Debug("Got request to update the state for teamcity log")
			// TODO: Make database support TeamCityData struct
			/*	if err := c.database.Put(tcdata, buildID); err != nil {
					logrus.WithField("build_id", buildID).WithError(err).Error("unable to add to database")
				}
				logrus.WithField("build_id", buildID).Info("database updated")
			*/
			c.logChannel <- tcdata.Logs
			logrus.Debug("build logs sent to tc.logChannel")
			logrus.WithFields(logrus.Fields{
				"build_id": buildID,
				"status":   tcdata.Status,
			}).Info("build finished")
			return
		}
	}
}

func (c *TeamCityCrawler) crawlBuild(uri *url.URL) {
	// TODO: check for build log size before downloading it, can be very big
	// Get buildLog metadata
	buildID := <-c.processChannel
	buildInfoUri := fmt.Sprintf("https://teamcity.local:8080/app/rest/builds/id:%s?fields=id,buildTypeId,number,status,startDate,finishDate", buildID)
	tcd := &ale.TeamCityData{}
	infoRequest, err := http.NewRequest("GET", buildInfoUri, nil)
	infoRequest.Header.Set("Authorization", "Bearer "+c.token)
	infoResp, err := c.httpClient.Do(infoRequest)
	if err != nil {
		logrus.Error(err)
	}
	defer infoResp.Body.Close()
	infoBody, err := ioutil.ReadAll(infoResp.Body)
	err = json.Unmarshal(infoBody, &tcd)

	// Get buildLog
	request, err := http.NewRequest("GET", uri.String(), nil)
	request.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(request)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Info(resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// Extract teamcity log data
	tcdata := c.extractBuildLogs(tcd, string(body), buildID)
	c.stateChannel <- tcdata
	logrus.Debug("teamcity data sent to stateChannel")
	return
}

func (c *TeamCityCrawler) extractBuildLogs(tcd *ale.TeamCityData, buildLog string, buildID string) *ale.TeamCityData {
	return &ale.TeamCityData{
		Status:      tcd.Status,
		BuildID:     buildID,
		BuildTypeID: tcd.BuildTypeID,
		Number:      tcd.Number,
		StartDate:   tcd.StartDate,
		FinishDate:  tcd.FinishDate,
		Logs:        c.splitLogs(buildLog),
	}
}

func (c *TeamCityCrawler) splitLogs(log string) []*ale.Log {
	var l []*ale.Log
	for _, line := range strings.Split(log, "\n") {
		if line == "" {
			continue
		}
		l = append(l, c.extractTimestamp(line))
	}
	return l
}

func (c *TeamCityCrawler) extractTimestamp(line string) *ale.Log {
	result := c.r.FindStringSubmatch(line)
	if result == nil {
		return &ale.Log{
			Line: line,
		}
	}
	return &ale.Log{
		TimeStamp: result[1],
		Line:      result[2],
	}
}
