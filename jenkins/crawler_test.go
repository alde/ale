package jenkins

import (
	"fmt"
	"testing"

	"github.com/alde/ale"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractTimestamp(t *testing.T) {
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
	}

	for i, td := range tdata {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := extractTimestamp(td.input)
			assert.Equal(t, td.expected, actual)
		})
	}
}
