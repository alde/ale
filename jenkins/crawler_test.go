package jenkins

import (
	"fmt"
	"testing"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/mock"
	"github.com/stretchr/testify/assert"
)

var (
	c = NewCrawler(&mock.DB{}, config.DefaultConfig())
)

func Test_ExtractTimestamp(t *testing.T) {
	t.Skip()
	tdata := []struct {
		input    string
		expected *ale.Log
	}{
		{
			`<span class="timestamp"><b>15:06:30</b> </span><style>.timestamper-plain-text {visibility: hidden;}</style>[WS-CLEANUP] done`,
			&ale.Log{
				TimeStamp: "15:06:30",
				Line:      "[WS-CLEANUP] done",
			},
		},
		{
			`a log line without an identifiable timestamp`,
			&ale.Log{
				Line: "a log line without an identifiable timestamp",
			},
		},
		{
			`<span class="timestamp"><b>15:06:30</b> </span><style>.timestamper-plain-text {visibility: hidden;}</style>[WS-CLEANUP] done`,
			&ale.Log{
				TimeStamp: "15:06:30",
				Line:      "[WS-CLEANUP] done",
			},
		},
	}

	for i, td := range tdata {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := c.extractTimestamp(td.input)
			assert.Equal(t, td.expected, actual)
		})
	}
}

func Test_SplitLogs(t *testing.T) {
	input := "<span class=\"timestamp\"><b>15:38:12</b> </span><span style=\"display: none\">[2019-02-14T15:38:12.376Z]</span> [WS-CLEANUP] Deleting project workspace...\n<span class=\"timestamp\"><b>15:38:12</b> </span><span style=\"display: none\">[2019-02-14T15:38:12.376Z]</span> [WS-CLEANUP] Deferred wipeout is used...\n<span class=\"timestamp\"><b>15:38:12</b> </span><span style=\"display: none\">[2019-02-14T15:38:12.381Z]</span> [WS-CLEANUP] done\n"
	expected := []*ale.Log{
		{
			TimeStamp: "2019-02-14T15:38:12.376Z",
			Line:      "[WS-CLEANUP] Deleting project workspace...",
		},
		{
			TimeStamp: "2019-02-14T15:38:12.376Z",
			Line:      "[WS-CLEANUP] Deferred wipeout is used...",
		},
		{
			TimeStamp: "2019-02-14T15:38:12.381Z",
			Line:      "[WS-CLEANUP] done",
		},
	}
	actual := c.splitLogs(input)
	assert.Equal(t, expected, actual)
}
