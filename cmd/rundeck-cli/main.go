package main

import (
	"log"
	"os"

	"github.com/ray1729/rundeck-cli/pkg/command"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "rundeck-cli"
	app.Version = "0.1.0"
	app.Usage = "Command-line interface to Rundeck"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "api-version",
			Usage:  "Rundeck API version",
			Value:  24,
			EnvVar: "RUNDECK_API_VERSION",
		},
		cli.StringFlag{
			Name:   "server-url",
			Usage:  "Rundeck server URL",
			EnvVar: "RUNDECK_SERVER",
		},
		cli.StringFlag{
			Name:   "rundeck-user",
			Usage:  "Rundeck username",
			EnvVar: "RUNDECK_USER,USER",
		},
		cli.StringFlag{
			Name:   "rundeck-password",
			Usage:  "Rundeck password",
			EnvVar: "RUNDECK_PASSWORD",
			Hidden: true,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "list-jobs",
			Usage:  "list the jobs that exist for a project",
			Action: command.ListJobs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name",
				},
			},
		},
		{
			Name:   "run-job",
			Usage:  "run a job specified by ID",
			Action: command.RunJob,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "id",
					Usage: "job ID",
				},
				cli.BoolFlag{
					Name:  "wait",
					Usage: "wait for job to complete and show status",
				},
				cli.BoolFlag{
					Name: "tail",
					Usage: "tail job output (implies --wait)",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
