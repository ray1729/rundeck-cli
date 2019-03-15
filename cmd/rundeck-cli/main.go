package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ray1729/rundeck-cli/pkg/client"
)

func main() {
	c, err := client.New(client.RundeckParams{
		APIVersion: 24,
		ServerUrl: os.Getenv("RUNDECK_SERVER"),
		Username: os.Getenv("RUNDECK_USER"),
		Password: os.Getenv("RUNDECK_PASSWORD"),
	})

	if err != nil {
		log.Fatal(err)
	}

	err = c.Login()
	if err != nil {
		log.Fatal(err)
	}

	res, err := c.ListJobs("Tech-Ops", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)


	helloWorld := "4a05c44d-2864-450a-a2a3-50fb3d5bd553"
	res2, err := c.RunJob(helloWorld, map[string]string{"name": "Jenkins"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res2)
}
