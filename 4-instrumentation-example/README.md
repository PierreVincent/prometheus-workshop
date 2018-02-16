# Sample application to instrument

This is a small Golang application, used to illustrate how to instrument code with built-in and custom Prometheus metrics. Follow instructions as part of [Part 4](../4-instrumenting-code.md)

## Requirements

- Golang (1.8+)

## Run the application

    go run *.go

By default, this will start an HTTP API on http://localhost:8888
    
## Use the application


This sample application simulates a Job Queue. It takes in Jobs and has workers process them in the background.

Jobs are pushed to the application using the API, e.g. 
    
    curl -X POST http://localhost:8888/jobs
    
_Or use the script [queueUpJobs.sh](queueUpJobs.sh) to push 100 jobs on the queue at once._

Each job will be processed by a background worker, between 5 and 10 seconds. If there are no jobs running, workers will back-off for a while before picking up new jobs.

### Pushing a new job

Requesting the endpoint `POST /jobs` will add a new Job on the queue. There is simple  which will put in 100 jobs on the queue.

    