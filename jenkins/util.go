package jenkins

import "strings"

// JobURLToAPI converts a Jenkins Job URL to it's API url
func JobURLToAPI(url string) string {
	return strings.Join([]string{strings.TrimRight(url, "/"), "wfapi", "describe"}, "/")
}

func ExtcractLogs() {

}
