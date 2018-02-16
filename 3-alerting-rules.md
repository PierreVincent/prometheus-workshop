# Part 3: Alerting rules

Grafana dashboards are a great place to start investigating issues, and so is the Prometheus UI for exploring single queries. This is however not enough to react to issues as they occur. Luckily, it's possible to define Prometheus rules using queries and thresholds, and alert if something abnormal is detected.

## Viewing rules and alert status

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

## Alerting on error rate

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