# Starting Prometheus, Alertmanager & Grafana on Kubernetes

Apply the full directory to your Kubernetes cluster (make sure to have the correct context active):

```
kubectl apply -f .
```

Check the status of the pods:

```
kubectl get pods -n prom-workshop
```

List exposed services:

```
kubectl get services -n prom-workshop
```

If you are using Kubernetes on a cloud provider which supports LoadBalancers, it will provide external IPs automatically. You should see an IP address under `EXTERNAL-IP` (it may be marked as `<pending>` for a few minutes). You can combine this with the relevant `PORT` to access each of the exposed services:

- prometheus (9090)
- grafana (3000)
- alertmanager (9093)
- http-simulator (8080)

This workshop assumes that your cloud provider supports LoadBalancers. If it's not the case , you will need to use the provisioned NodePorts, displayed after the `:` in the `PORT(S)` column.

## Cleanup

```
kubectl delete -f .
```