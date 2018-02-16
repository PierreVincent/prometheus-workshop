# Part 2: Building Grafana Dashboards

Querying Prometheus on a ad-hoc basis is useful to explore data, especially when troubleshooting issues. In the longer term, some queries might be very useful to keep looking at, especially as a starting point to investigating production problems.

Grafana makes it very easy to transition from single graphs to very slick dashboards, so we're going to look at combining the interesting data we have from the queries in Part 1.

## First visit to Grafana

Navigate the Grafana URL (provided during workshop - or your own Grafana)

Login with default credentials: `admin` / `admin`

Navigate to Data Sources: these are the time-series Grafana will query for graphs. Grafana itself does not store any metric, only the dashboards configurations.

For this workshop, there is only one Data Source, which is the Prometheus that we used previously.

## Starting a new dashboard

Navigate to _Dashboards > New_ to create a new blank dashboard. Save the dashboard under your name so that everybody can work on their own, _e.g. `Pierre Vincent Workshop`_. 

Using drag-and-drop, you can create new widgets on the blank row. The next sections go over some of the metrics that were explored with the Prometheus UI, and how we can add them as widgets in this new dashboard.

## Graphing request rate

Let's create our first widget by dragging a new Graph panel to the blank row. Click on _Panel Title_ and _Edit_ to customise this new graph. The tabs at the bottom display all the options for this graph.

Navigate to _General_ and edit the title of this graph to “Request Rate”.

Navigate to _Metrics_ and add the request rate query:

    sum(rate(http_requests_total{app="http-simulator"}[5m]))

The legend is currently `{}`, you can update this under the metric query, to “request/s”.

You can use the other tabs to further customise the graph:
- _Axis_ to specify units and annotate different axes
- _Legend_ to format as a table, display min/max/current values for each series
- _Display_ to play with look-and-feel of the graph

## Displaying live error rate

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

## Table of Top-requested endpoints

Create a new row and drag a new Table panel.

Add the following metric (request rate for each endpoint):

    sum(rate(http_requests_total{app="http-simulator"}[5m])) by (endpoint)

Because Prometheus is returning time series for the entire time period, the table contains way too many values. Under _Time Range_, set _Override relative time_ to _Last_ `1s` and check _Hide time override info_. There should only be 5 entries displayed in the table.

The time column is not relevant so we can hide it under _Column Styles_. We can also rename the `Value` column to something more meaningful:

- Add a Column Style
- Set _Apply to columns named_ to `Value`
- Set _Column Header_ to `Requests/s`

Finally, click on the `Requests/s` header in the table to order the table by most active endpoints first.

## Get creative!

It's your turn now! You can take the queries from the previous section, or come up with your own. Try and enrich this dashboard with information that you think would be useful to monitor this service.

You can look at other metrics available for the service under URL_OF_SERVICE_METRICS

If you're looking for ideas:

- Graph of latency distribution
- Cumulative % graph of endpoint request rate
- Memory usage over time
- CPU usage over time
- Graph % of requests fulfilling the SLO of 400ms for /login endpoint