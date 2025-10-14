package robots

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Robots struct {
	rules map[string][]string
}

func (r *Robots) IsAllowed(userAgent string, u *url.URL) bool {
	currentRules, ok := r.rules[userAgent]
	if !ok {
		if currentRules, ok = r.rules["*"]; !ok {
			return true
		}
	}
	for _, rule := range currentRules {
		if strings.Contains(u.String(), rule) {
			return false
		}
	}
	return true
}

func FetchRobots(domain string) (*Robots, error) {
	u := url.URL{
		Scheme: "http",
		Host:   domain,
		Path:   "robots.txt",
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	errClose := resp.Body.Close()
	if errClose != nil {
		return nil, errClose
	}

	if resp.StatusCode == http.StatusNotFound {
		return &Robots{nil}, nil
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	rules := make(map[string][]string)
	userAgent := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "User-agent:") {
			line = strings.TrimSpace(line)
			userAgent = strings.TrimPrefix(line, "User-agent: ")
			continue
		}
		if strings.Contains(line, "Disallow:") {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "Disallow: ")
			fields := strings.Fields(line)
			for _, field := range fields {
				rules[userAgent] = append(rules[userAgent], field)
			}
		}
	}
	return &Robots{rules}, nil
}
