# Rundeck CLI

A basic (and incomplete) command-line interface for Rundeck.

## Installation

    go get -u github.com/ray1729/rundeck-cli

## Usage

    main [global options] command [command options] [arguments...]

## Commands

     list-jobs         list the jobs that exist for a project
     execution-output  Dump the output for the specified execution
     execution-info    Dump the execution info for the specified execution
     run-job           run a job specified by ID
     help, h           Shows a list of commands or help for one command

## Global Options

    --api-version value   Rundeck API version (default: 24) [$RUNDECK_API_VERSION]
    --server-url value    Rundeck server URL [$RUNDECK_SERVER]
    --rundeck-user value  Rundeck username [$RUNDECK_USER, $USER]
    --help, -h            show help
    --version, -v         print the version

## Author

Ray Miller <ray@1729.org.uk>
