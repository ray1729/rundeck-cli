package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	defaultAPIVersion = 30
)

type RundeckParams struct {
	APIVersion                    int
	ServerUrl, Username, Password string
}

type Client struct {
	*RundeckParams
	http *http.Client
}

func New(s RundeckParams) (*Client, error) {
	if strings.HasSuffix(s.ServerUrl, "/") {
		s.ServerUrl = s.ServerUrl[0 : len(s.ServerUrl)-1]
	}
	if len(s.ServerUrl) == 0 {
		return nil, fmt.Errorf("Rundeck client requires ServerUrl")
	}
	if len(s.Username) == 0 || len(s.Password) == 0 {
		return nil, fmt.Errorf("Rundeck client requires Username and Password")
	}

	if s.APIVersion == 0 {
		s.APIVersion = defaultAPIVersion
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpClient := http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Path == "/user/login" || req.URL.Path == "/user/error" {
				return fmt.Errorf("Authentication error")
			}
			return nil
		},
	}

	return &Client{&s, &httpClient}, nil
}

func (c *Client) Login() error {
	resp, err := c.http.PostForm(
		c.ServerUrl+"/j_security_check",
		url.Values{"j_username": {c.Username}, "j_password": {c.Password}},
	)
	if err != nil {
		return fmt.Errorf("Login failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Login failed: unexpected HTTP status %s", resp.Status)
	}
	return nil
}

func (c *Client) ApiUrl(components ...string) string {
	apiUrl := fmt.Sprintf("%s/api/%d", c.ServerUrl, c.APIVersion)
	for _, component := range components {
		apiUrl = apiUrl + "/" + component
	}
	return apiUrl
}

type ListJobsResponse []struct{
	ID              string `json:"id"`
	Name            string `json:"name"`
	Group           string `json:"group"`
	Project         string `json:"project"`
	Href            string `json:"href"`
	Permalink       string `json:"permalink"`
	Scheduled       bool   `json:"scheduled"`
	ScheduleEnabled bool   `json:"scheduleEnabled"`
	Enabled         bool   `json:"enabled"`
}

func (c *Client) ListJobs(project string, params map[string]string) (ListJobsResponse, error) {
	req, err := http.NewRequest("GET", c.ApiUrl("project", project, "jobs"), nil)
	if err != nil {
		return nil, err
	}
	if len(params) != 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Add("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ListJobsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) RunJob(jobId string, options map[string]string) (string, error) {
	body, err := json.Marshal(map[string]map[string]string{"options": options})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ApiUrl("job", jobId, "executions"), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
