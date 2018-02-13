# Workshop: Cloud Native Monitoring with Prometheus & Grafana

## Objectives

This workshop covers:

- The core concepts of Prometheus
- Querying Prometheus to gain insights on how your applications behave
- Building Grafana dashboards combining multiple metrics
- Defining rules to trigger alerts based on metrics and thresholds
- Instrumenting Golang code to expose built-in and custom metrics

## Part 1: Querying Prometheus

### Requirements

If you are doing the workshop live, you will be given access to a temporary Prometheus & Grafana on which we will do the exercises, all you will need is a web browser.
 
If you are doing this at home, you can refer the [kubernetes/install](kubernetes/install) directory to start your own setup in a Kubernetes cluster.

### A first look at the Prometheus UI

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

### Request Rate

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

#### Exercices

- What's the overall request rate (with a 1 minute rolling-window) for the http-simulator service?
- How many requests per minute are errors?
- What's the error rate (in %) of requests to the /users endpoint?

### Latency distribution

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

#### Exercises:

- What is the median latency of all requests to the http-simulator service?
- What is the 90-percentile latency of all error requests to the http-simulator service, for each endpoint?
- Does the `/users` endpoint fulfill the SLO of _3 Nines_ requests responding within 400ms?

## Part 2: Building Grafana Dashboards

Querying Prometheus on a ad-hoc basis is useful to explore data, especially when troubleshooting issues. In the longer term, some queries might be very useful to keep looking at, especially as a starting point to investigating production problems.

Grafana makes it very easy to transition from single graphs to very slick dashboards, so we're going to look at combining the interesting data we have from the queries in Part 1.

### First visit to Grafana

Navigate the Grafana URL (provided during workshop - or your own Grafana)

Login with default credentials: `admin` / `admin`

Navigate to Data Sources: these are the time-series Grafana will query for graphs. Grafana itself does not store any metric, only the dashboards configurations.

For this workshop, there is only one Data Source, which is the Prometheus that we used previously.

### Starting a new dashboard

Navigate to _Dashboards > New_ to create a new blank dashboard. Save the dashboard under your name so that everybody can work on their own, _e.g. `Pierre Vincent Workshop`_. 

Using drag-and-drop, you can create new widgets on the blank row. The next sections go over some of the metrics that were explored with the Prometheus UI, and how we can add them as widgets in this new dashboard.

### Graphing request rate

Let's create our first widget by dragging a new Graph panel to the blank row. Click on _Panel Title_ and _Edit_ to customise this new graph. The tabs at the bottom display all the options for this graph.

Navigate to _General_ and edit the title of this graph to “Request Rate”.

Navigate to _Metrics_ and add the request rate query:

    sum(rate(http_requests_total{app="http-simulator"}[5m]))

The legend is currently `{}`, you can update this under the metric query, to “request/s”.

You can use the other tabs to further customise the graph:
- _Axis_ to specify units and annotate different axes
- _Legend_ to format as a table, display min/max/current values for each series
- _Display_ to play with look-and-feel of the graph

### Displaying live error rate

While data like error rate can be interesting to look over time, having a single figure for the current error rate can help give “at-a-glance” health check for the service.

Create new row and add add a new _Singlestat_ widget, and add the overall service error rate query:

    sum(rate(http_requests_total{app="http-simulator", status="500"}[5m])) / sum(rate(http_requests_total{app="http-simulator"}[5m]))

This should show some number like 0.010. The number is a percentage between 0 and 1, so we can change the unit under _Options>Unit_ (set to _None>percent(0.0-1.0)_)

The current value displayed is also the average over the Dashboard time window, instead we want to see the value at the end of the time-window (i.e. now). You can change this under Options>Stat (set to _Current_).

We can now add thresholds and colors to use the error rate as a health signal for the service. Under Options>Coloring, check _Value_ and set _Threshold_ to `0.01,0.05`, which means:

- Green: 0-1%
- Orange: 1-5%
- Red: >5%

The status can even be displayed as a gauge to indicate the thresholds. Under Options>Gauge, check _Show_ and set _Max_ to `1`.

It's also possible to overlay the evolution over the Grafana time period, by enabling the _Spark lines_.

### Table of Top-requested endpoints

Create a new row and drag a new Table panel.

Add the following metric (request rate for each endpoint):

    sum(rate(http_requests_total{app="http-simulator"}[5m])) by (endpoint)

Because Prometheus is returning time series for the entire time period, the table contains way too many values. Under _Time Range_, set _Override relative time_ to _Last_ `1s` and check _Hide time override info_. There should only be 5 entries displayed in the table.

The time column is not relevant so we can hide it under _Column Styles_. We can also rename the `Value` column to something more meaningful:

- Add a Column Style
- Set _Apply to columns named_ to `Value`
- Set _Column Header_ to `Requests/s`

Finally, click on the `Requests/s` header in the table to order the table by most active endpoints first.

### Get creative!

It's your turn now! You can take the queries from the previous section, or come up with your own. Try and enrich this dashboard with information that you think would be useful to monitor this service.

You can look at other metrics available for the service under URL_OF_SERVICE_METRICS

If you're looking for ideas:

- Graph of latency distribution
- Cumulative % graph of endpoint request rate
- Memory usage over time
- CPU usage over time
- Graph % of requests fulfilling the SLO of 400ms for /login endpoint

## Part 3: Alerting rules

Grafana dashboards are a great place to start investigating issues, and so is the Prometheus UI for exploring single queries. This is however not enough to react to issues as they occur. Luckily, it's possible to define Prometheus rules using queries and thresholds, and alert if something abnormal is detected.

### Viewing rules and alert status

Navigate to _Status > Rules_ to see the rules currently defined.

Rules are composed of several parts:
- `alert` is the name of the alert
- `expr` is the alert predicate (using Prometheus queries)
- `for` is the time to wait until triggering the alert
- `labels` is a map of additional labels for the alert

Let's look at the alert called `HttpSimulatorDown`:

    alert: HttpSimulatorDown
    expr: sum(up{app="http-simulator"}) == 0
    for: 1m
    labels:
      severity: critical
     
This alert will trigger if the total number of instance of `http-simulator` remains at 0 for at least 1 minute, and the alert will be labeled as `{severity="critical"}` 

Navigate to _Alerts_ to see the state of the current alerting rules (everything should be green)

### Alerting on error rate

_This part requires your own setup in a Kubernetes cluster. See [kubernetes/install](kubernetes/install) for details._

Rules are configured as yaml files, pointed to in the Prometheus configuration. If you are using your own Prometheus setup, you can add rules my modifying the ConfigMap `prom-workshop/prometheus-rules` (see [kubernetes/install/3-prometheus.yaml](kubernetes/install/3-prometheus.yaml)). You can edit this ConfigMap on the fly:

    kubectl edit configmap prometheus-rules -n prom-workshop

There is a single yaml file defined in the `data:` field. Edit this yaml snippet to add an entry to the `rules:` array.

    - alert: ErrorRateHigh
      expr: sum(rate(http_requests_total{app="http-simulator", status="500"}[5m])) / sum(rate(http_requests_total{app="http-simulator"}[5m])) > 0.02
      for: 1m
      labels:
        severity: major

Save the ConfigMap.

Prometheus does not reload rules automatically, to do so you need to hit the reload endpoint:

    curl -X POST http://PROMETHEUS_URL:9090/-/reload
    
The rule should now be displayed under _Config > Rules_ and _Alerts_.

## Part 4: Instrumentation

WIP

## Additional resources

- [Prometheus documentation](https://prometheus.io/docs/introduction/overview/)
- Prometheus Blog Series:
    - [Metrics and Labels](https://pierrevincent.github.io/2017/12/prometheus-blog-series-part-1-metrics-and-labels/)
    - [Metric Types](https://pierrevincent.github.io/2017/12/prometheus-blog-series-part-2-metric-types/)
    - [Exposing and collecting metrics](https://pierrevincent.github.io/2017/12/prometheus-blog-series-part-3-exposing-and-collecting-metrics/)
    - [Instrumenting code in Go and Java](https://pierrevincent.github.io/2017/12/prometheus-blog-series-part-4-instrumenting-code-in-go-and-java/)
    - [Alerting rules](https://pierrevincent.github.io/2017/12/prometheus-blog-series-part-5-alerting-rules/)
- [Robust Perception Blog](https://www.robustperception.io/blog/)
- [Prometheus: Best Practices and Beastly Pitfalls](https://www.youtube.com/watch?v=_MNYuTNfTb4)  by Julius Voltz (PromCon 2017)