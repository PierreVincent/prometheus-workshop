# Part 4: Instrumenting Code

In this part, we will be using the sample application in [instrumentation-example](instrumentation-example).

## Adding built-in Go metrics

The first thing to do for exposing Prometheus metrics is to add the `/metrics` endpoint. Using the [Prometheus Go Client](https://github.com/prometheus/client_golang), it will instrument the application with built-in Go metrics.

In `main.go`:
- Import `github.com/prometheus/client_golang/prometheus/promhttp`
- Add the metrics handler `router.Handle("/metrics", promhttp.Handler())`

Start the service:

    go run *.go
    
The default metrics should now be accessible on [http://localhost:8888/metrics](http://localhost:8888/metrics)

## Adding custom metrics

Default Go metrics are interesting for low-level monitoring but the real value of instrumentation comes from exposing the internal state of the application.

### Current queue size

In `jobs.go`, create a new Gauge:

```go
var jobQueueSize = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "job_queue_size",
        Help: "Current number of jobs waiting in queue",
    },
)
```
    
Register this new metric (it won't show up in `/metrics` otherwise):

```go
func init() {
    prometheus.MustRegister(jobQueueSize)
}
```
 
Increment the Gauge in `Push()`:

```go
func (jq *JobQueue) Push(job *Job) {
    jq.mutex.Lock()
    defer jq.mutex.Unlock()

    jq.q = append(jq.q, job)
    jobQueueSize.Inc()
}
```

Decrement the Gauge in `Pull()`;

```go
func (jq *JobQueue) Pull() (job *Job) {
    jq.mutex.Lock()
    defer jq.mutex.Unlock()

    if len(jq.q) > 0 {
        job = jq.q[0]
        jq.q = jq.q[1:]
        jobQueueSize.Dec()
    }
    return
}
```

Start the service and push some jobs in the queue. You should now see the `job_queue_size` metric change over time, as jobs get completed.

### Distribution of job duration

In `worker.go`, create a new Histogram:

```go

var jobsCompletionDurationSeconds = prometheus.NewHistogram(
    prometheus.HistogramOpts{
        Name: "jobs_completion_duration_seconds",
        Help: "Histogram of job completion time",
        Buckets: prometheus.LinearBuckets(4, 1, 16),
    },
)
```

Don't forget to register the new metric:

```go
func init() {
    prometheus.MustRegister(jobsCompletionDurationSeconds)
}
```

Instrument the `pullJobAndRun()` function:

```go
func (w *Worker) pullJobAndRun() {
    job := w.queue.Pull()
    if job != nil {
        jobStart := time.Now()
        log.Printf("[Worker %s] Starting job: %s", w.id, job.ID)
        job.run()
        log.Printf("[Worker %s] Finished job: %s", w.id, job.ID)
        jobsCompletionDurationSeconds.Observe(time.Since(jobStart).Seconds())
    } else {
        log.Printf("[Worker %s] Queue is empty. Backing off 5 seconds", w.id)
        time.Sleep(5 * time.Second)
    }
}
```

### More metrics!

Here are a few other things that could be worth instrumenting: 

- Total number of completed jobs 
- Current number of active jobs
- Current number of active workers
- Current number of idle workers
