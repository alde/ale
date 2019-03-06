package ale

// Link represents a relative uri deeper into the Jenkins API
type Link struct {
	Href string `json:"href"`
}

// JobData holds parts of a jenkins job
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

// JobStage holds information about a stage of a job
type JobStage struct {
	Links struct {
		Self Link `json:"self"`
	} `json:"_links"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
}

// JobExecution holds information regarding an execution of a job
type JobExecution struct {
	Links struct {
		Self Link `json:"self"`
		Log  Link `json:"log"`
	} `json:"_links"`
	ID              string          `json:"id"`
	Status          string          `json:"status"`
	Name            string          `json:"name"`
	StartTimeMillis int             `json:"startTimeMillis"`
	StageFlowNodes  []StageFlowNode `json:"stageFlowNodes"`
}

// StageFlowNode holds information regarding a flow-node in a stage
type StageFlowNode struct {
	Links struct {
		Self Link `json:"self"`
		Log  Link `json:"log"`
	} `json:"_links"`
	ID                   string `json:"id"`
	Status               string `json:"status"`
	Name                 string `json:"name"`
	StartTimeMillis      int    `json:"startTimeMillis"`
	ParameterDescription string `json:"parameterDescription"`
}

// NodeLog maps to the logs from a node
type NodeLog struct {
	NodeID     string `json:"nodeId"`
	NodeStatus string `json:"nodeStatus"`
	Length     int    `json:"length"`
	HasMore    bool   `json:"hasMore"`
	Text       string `json:"text"`
	ConsoleURL string `json:"consoleUrl"`
}

// JenkinsData is the topmost level of the flattened structure stored in the database
type JenkinsData struct {
	Stages  []*JenkinsStage `json:"stages"`
	Status  string          `json:"status"`
	Name    string          `json:"name"`
	ID      string          `json:"id"`
	BuildID string          `json:"build_id"`
}

// JenkinsStage holds the output from a given stage
type JenkinsStage struct {
	Status      string          `json:"status"`
	Name        string          `json:"name"`
	Logs        []*Log          `json:"log"`
	LogLength   int             `json:"log_length"`
	SubStages   []*JenkinsStage `json:"substage"`
	StartTime   int             `json:"start_time"`
	Description string          `json:"description"`
}

// The Log struct maps to the response value for the structured log
type Log struct {
	TimeStamp string `json:"timestamp"`
	Line      string `json:"line"`
}

// DatastoreEntity is used to store data in datastore, and prevent indexing of the huge json
type DatastoreEntity struct {
	Key   string      `json:"key" datastore:"key"`
	Value JenkinsData `json:"value" datastore:"value,noindex"`
}
