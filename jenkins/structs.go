package jenkins

type Link struct {
	Href string `json:"href"`
}

type JobData struct {
	Links struct {
		Self      Link `json:"self"`
		Artifacts Link `json:"artifacts"`
	} `json:"_links"`
	Stages []JobStage `json:"stages"`
	Status string     `json:"status"`
	Name   string     `json:"name"`
	ID     string     `json:"id"`
}

type JobStage struct {
	Links struct {
		Self Link `json:"self"`
	} `json:"_links"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
}

type JobExecution struct {
	Links struct {
		Self Link `json:"self"`
		Log  Link `json:"log"`
	} `json:"_links"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
}

type JobExecutionLog struct {
	NodeID     string `json:"nodeId"`
	NodeStatus string `json:"nodeStatus"`
	Length     int    `json:"length"`
	HasMore    bool   `json:"hasMore"`
	Text       string `json:"text"`
	ConsoleUrl string `json:"consoleUrl"`
}
