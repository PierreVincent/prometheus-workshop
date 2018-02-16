# Part 1: Querying Prometheus

## Requirements

If you are doing the workshop live, you will be given access to a temporary Prometheus & Grafana on which we will do the exercises, all you will need is a web browser.
 
If you are doing this at home, you can refer the [kubernetes/install](kubernetes/install) directory to start your own setup in a Kubernetes cluster.

## A first look at the Prometheus UI

Navigate to the Prometheus URL (provided during workshop - or your own Prometheus).

You should see the Prometheus UI, with the following tabs:
- Alerts: list of alerting rules and their status
- Graph: default view, where we can you can query time-series
- Status: current state of scraped targets, service discovery as well as various configuration
- Help: link to the Prometheus official documentation (essential!)

Before we start querying time-series, navigate to _Status > Targets_. This page displays all the different targets that this Prometheus service is currently scraping. Each target has a given state `UP` or `DOWN`, which indicates whether the _last scrape_ attempt by Prometheus was successful.

Prometheus also exposes this state as the `up` metric. Navigate back to the default Graph view to query it:

    up

The result is a list of time-series with their latest value of either 0 and 1, for each target scraped by Prometheus.

## Request Rate

One of the targets we will be working with in this workshop is a dummy target `http-simulator`. This service exposes similar metrics as a real HTTP API, and simulated various levels of activity, latency and errors.

First let's make sure that this target is working correctly:

    up{app="http-simulator"}
    
The `{name="value"}` operator filters the `up` metrics for a specific label.

Now let's look at the number of requests that this service has received since it started up:

    http_requests_total{app="http-simulator"}

Results show a lot of different time series, with a lot of different labels. Can you tell what the difference is between each of these time series?

Each time-series in the list is for the same `http_requests_total` metric, but with different label sets. This is very powerful, because now we're able to make the distinction between different endpoints and their status code.

Let's imagine you're only interested in the number of successful requests for the `/login` endpoint. To achieve this, you can add more label filters:

    http_requests_total{app="http-simulator", status="200", endpoint="/login"}

Now we may also be interested in the total number of successful requests, regardless of the endpoint:

    http_requests_total{app="http-simulator", status="200"}

This is still returning multiple time-series, but we can aggregate them using the `sum()` function:

    sum(http_requests_total{app="http-simulator", status="200"})

This is interesting, but does not exactly tell us much in terms of request rate. The graph just keeps going up. If instead we want to look at the number of requests per second, we can use the `rate()` function.

    sum(rate(http_requests_total{app="http-simulator", status="200"}[5m]))

We should now have a single time-series, displaying the number of successful requests per second. The `[5m]` operator used inside the `rate()` function means that the rate is calculated over a 5 minutes rolling-window. The shorter this period, the more _spikey_ the graph will be. You need this to be greater than your scrape interval.

### Exercises

- What's the overall request rate (with a 1 minute rolling-window) for the http-simulator service?
- How many requests per minute are errors?
- What's the error rate (in %) of requests to the /users endpoint?

## Latency distribution

Latency cannot be accurately measured through sums and averages, because it's a distribution problem. The `http-simulator` exposes an _Histogram_ metric for this purpose, called `http_request_duration_milliseconds`. Histogram metrics are actually multiple metrics, with the `_buckets` suffix.

    http_request_duration_milliseconds_bucket{app="http-simulator"}

Again, this metric is labeled by endpoint and status, so let's look at only the successful login requests for now:

    http_request_duration_milliseconds_bucket{service="http-simulator", status="200", endpoint="/login"}

The remaining time series should only differ by their `le` label, which stands for "Less than or equal". For example, the counter with the label `le=50` is the number of successful login requests that took at most 50ms to complete.

These buckets are predefined by the service when it is instrumented, so it's important to have a good understanding of the latency profile of the service so that they are meaningful.

In addition of buckets, Histograms expose 2 other metrics:
- `_sum` (total sum of observed values)
- `_count` (total count of observed values)

These can be useful to derive rates out of the buckets. Let's imagine our login SLO is that 99% of requests respond within 200ms, we can try to query for the % of login requests within the SLO.

    http_request_duration_milliseconds_bucket{service="http-simulator", status="200", endpoint="/login", le="200"} / http_request_duration_milliseconds_count{service="http-simulator", status="200", endpoint="/login"}

Another approach is to query Prometheus for the actual 99-percentile, using the `histogram_quantile()` function:

    histogram_quantile(0.99, rate(http_request_duration_milliseconds_bucket{app="http-simulator", status="200", endpoint="/login"}[5m]))

### Exercises

- What is the median latency of all requests to the http-simulator service?
- What is the 90-percentile latency of all error requests to the http-simulator service, for each endpoint?
- Does the `/users` endpoint fulfill the SLO of _3 Nines_ requests responding within 400ms?