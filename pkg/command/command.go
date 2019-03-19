package command

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/ray1729/rundeck-cli/pkg/rundeck"
	"gopkg.in/urfave/cli.v1"
)

func initClient(c *cli.Context) (*rundeck.Client, error) {
	client, err := rundeck.NewClient(rundeck.ClientParams{
		APIVersion: c.GlobalInt("api-version"),
		ServerUrl:  c.GlobalString("server-url"),
		Username:   c.GlobalString("rundeck-user"),
		Password:   c.GlobalString("rundeck-password"),
	})

	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to initialize Rundeck client: %v", err), 3)
	}

	err = client.Login()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Rundeck login failed: %v", err), 3)
	}

	return client, nil
}

func ListJobs(c *cli.Context) error {
	project := c.String("project")
	if len(project) == 0 {
		return cli.NewExitError("project is required", 3)
	}

	rd, err := initClient(c)
	if err != nil {
		return err
	}

	res, err := rd.ListJobs(project, nil)
	if err != nil {
		return cli.NewExitError(err.Error(), 4)
	}

	dumpJSON(res)

	return nil
}

func ExecutionOutput(c *cli.Context) error {
	executionId := c.Int("execution")
	if executionId == 0 {
		return cli.NewExitError("execution is required", 3)
	}

	rd, err := initClient(c)
	if err != nil {
		return err
	}

	res, err := rd.ExecutionOutput(executionId, "0")
	if err != nil {
		return err
	}

	dumpJSON(res)

	return nil
}

func ExecutionInfo(c *cli.Context) error {
	executionId := c.Int("execution")
	if executionId == 0 {
		return cli.NewExitError("execution is required", 3)
	}

	rd, err := initClient(c)
	if err != nil {
		return err
	}

	res, err := rd.ExecutionInfo(executionId)
	if err != nil {
		return err
	}

	dumpJSON(res)

	return nil
}

func RunJob(c *cli.Context) error {
	jobId := c.String("id")
	if len(jobId) == 0 {
		return cli.NewExitError("job ID is required", 3)
	}

	options := make(map[string]string)
	for _, s := range c.Args() {
		optval := strings.SplitN(s, "=", 2)
		if len(optval) != 2 {
			return cli.NewExitError(fmt.Sprintf("failed to parse option '%s'", s), 3)
		}
		options[optval[0]] = optval[1]
	}

	rd, err := initClient(c)
	if err != nil {
		return err
	}

	job, err := rd.RunJob(jobId, options)
	if err != nil {
		return cli.NewExitError(err.Error(), 4)
	}

	fmt.Printf("Submitted execution %d <%s>\n", job.Id, job.Href)

	if c.Bool("tail") {
		if err = tailOutput(rd, job.Id); err != nil {
			return cli.NewExitError(err.Error(), 4)
		}
	}

	if c.Bool("tail") || c.Bool("wait") {
		if err = awaitCompletion(rd, job.Id); err != nil {
			return cli.NewExitError(err.Error(), 4)
		}
	}

	return nil
}

func tailOutput(rd *rundeck.Client, executionId int) error {

	offset := "0"

	for {
		out, err := rd.ExecutionOutput(executionId, offset)
		if err != nil {
			return err
		}
		for _, e := range out.Entries {
			fmt.Println(e.Time + " " + e.Log)
		}
		if out.ExecCompleted {
			break
		}
		offset = out.Offset
		time.Sleep(2 * time.Second)
	}

	return nil
}

func awaitCompletion(rd *rundeck.Client, executionId int) error {
	var state *rundeck.ExecutionStateResponse
	n := 0.0
	for {
		var err error
		state, err = rd.ExecutionState(executionId)
		if err != nil {
			return err
		}
		if state.Completed {
			break
		}
		secs := math.Exp2(n)
		if secs > 30 {
			secs = 30
		}
		time.Sleep(time.Duration(secs) * time.Second)
		n++
	}

	if state.State != "SUCCEEDED" {
		return fmt.Errorf("Job execution failed")
	}

	return nil
}

func dumpJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(v); err != nil {
		log.Fatal("Failed to encode JSON: %v", err)
	}
}
