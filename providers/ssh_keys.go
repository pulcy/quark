package providers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// FetchSSHKeys uses a github account to fetch public SSH keys
func FetchSSHKeys(githubAccount string) ([]string, error) {
	if githubAccount == "" || githubAccount == "-" {
		return nil, nil
	}
	resp, err := http.Get(fmt.Sprintf("https://github.com/%s.keys", githubAccount))
	if err != nil {
		return nil, maskAny(err)
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, maskAny(err)
	}
	keys := []string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	return keys, nil
}
