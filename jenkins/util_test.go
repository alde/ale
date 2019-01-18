package jenkins

import (
	"fmt"
	"testing"
)

func Test_JobURLToAPI(t *testing.T) {
	testdata := []struct {
		input    string
		expected string
	}{
		{"http://jenkins.local:8080/job/jobId/262", "http://jenkins.local:8080/job/jobId/262/wfapi/describe"},
		{"http://jenkins.local:8080/job/jobId/262/", "http://jenkins.local:8080/job/jobId/262/wfapi/describe"},
	}

	for idx, td := range testdata {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			actual := JobURLToAPI(td.input)
			if actual != td.expected {
				t.Errorf("expected: %s\nactual: %s\n", td.expected, actual)
			}
		})
	}
}
