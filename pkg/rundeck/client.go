package rundeck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAPIVersion = 30
)

type ClientParams struct {
	APIVersion                    int
	ServerUrl, Username, Password string
}

type Client struct {
	*ClientParams
	http *http.Client
}

func NewClient(s ClientParams) (*Client, error) {
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

type ErrorResponse struct {
	Error      bool   `json:"error"`
	APIVersion int    `json:"apiversion"`
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message"`
}

func errorResponse(payload []byte) error {
	var response ErrorResponse
	err := json.Unmarshal(payload, &response)
	if err == nil && response.Error {
		return fmt.Errorf(response.Message)
	}
	return nil
}

type ListJobsResponse []struct {
	Id              string `json:"id"`
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

	if err = errorResponse(body); err != nil {
		return nil, fmt.Errorf("ListJobs failed: %v", err)
	}

	var result ListJobsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type DateTime struct {
	UnixTime int    `json:"unixtime"`
	Date     string `json:"date"`
}

type JobDetails struct {
	Id              string            `json:"id"`
	Name            string            `json:"name"`
	Group           string            `json:"group"`
	Project         string            `json:"project"`
	Description     string            `json:"Descrption"`
	AverageDuration int               `json:"avegareDuration"`
	Options         map[string]string `json:"options"`
	Href            string            `json:"href"`
	Permalink       string            `json:"permalink"`
}

type RunJobResponse struct {
	Id            int        `json:"id"`
	Href          string     `json:"href"`
	Permalink     string     `json:"permalink"`
	Status        string     `json:"status"`
	Project       string     `json:"project"`
	ExecutionType string     `json:"executionType"`
	User          string     `json:"user"`
	DateStarted   DateTime   `json:"date-started"`
	Job           JobDetails `json:"job"`
	Description   string     `json:"description"`
	ArgString     string     `json:"argstring"`
}

func (c *Client) RunJob(jobId string, options map[string]string) (*RunJobResponse, error) {
	body, err := json.Marshal(map[string]map[string]string{"options": options})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ApiUrl("job", jobId, "executions"), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = errorResponse(resBody); err != nil {
		return nil, fmt.Errorf("RunJob failed: %v", err)
	}

	var result RunJobResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type ExecutionInfoResponse struct {
	Id              int        `json:"id"`
	Href            string     `json:"href"`
	Permalink       string     `json:"permalink"`
	Status          string     `json:"status"`
	Project         string     `json:"project"`
	User            string     `json:"user"`
	DateStarted     DateTime   `json:"date-started"`
	DateEnded       DateTime   `json:"date-ended"`
	Job             JobDetails `json:"job"`
	Description     string     `json:"description"`
	ArgString       string     `json:"argstring"`
	SuccessfulNodes []string   `json:"successfulNodes"`
	FailedNodes     []string   `json:"failedNodes"`
}

func (c *Client) ExecutionInfo(id int) (*ExecutionInfoResponse, error) {
	req, err := http.NewRequest("GET", c.ApiUrl("execution", fmt.Sprintf("%d", id)), nil)
	if err != nil {
		return nil, err
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

	if err = errorResponse(body); err != nil {
		return nil, fmt.Errorf("ExecutionInfo failed: %v", err)
	}

	var result ExecutionInfoResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type ExecutionStateResponse struct {
	Id        int    `json:"executionId"`
	Error     string `json:"error"`
	Completed bool   `json:"completed"`
	State     string `json:"executionState"`
}

func (c *Client) ExecutionState(id int) (*ExecutionStateResponse, error) {
	req, err := http.NewRequest("GET", c.ApiUrl("execution", fmt.Sprintf("%d", id), "state"), nil)
	if err != nil {
		return nil, err
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

	if err = errorResponse(body); err != nil {
		return nil, fmt.Errorf("ExecutionState failed: %v", err)
	}

	var result ExecutionStateResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type ExecutionOutputResponse struct {}

func (c *Client) ExecutionOutput(id int, offset int) (*ExecutionOutputResponse, error) {
	req, err := http.NewRequest("GET", c.ApiUrl("execution", fmt.Sprintf("%d", id), "output"), nil)
	if err != nil {
		return nil, err
	}
	//req.Header.Add("Accept", "applicaton/json")

	if offset > 0 {
		q := req.URL.Query()
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
	}

	fmt.Println(req.URL.String())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)

	if resp.StatusCode == http.StatusNotFound {
		time.Sleep(500*time.Millisecond)
		return c.ExecutionOutput(id, offset)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ExecutionOutput returned HTTP status %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = errorResponse(body); err != nil {
		return nil, fmt.Errorf("ExecutonOutput failed: %v", err)
	}

	fmt.Println(string(body))

	var result ExecutionOutputResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
