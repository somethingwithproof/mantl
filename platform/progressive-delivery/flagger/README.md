# Flagger - Progressive Delivery

Automated canary deployments with metrics analysis and rollback.

## Features

- **Automated Canary Analysis**: Metrics-based promotion or rollback
- **Multiple Strategies**: Canary, A/B testing, Blue/Green
- **Gateway API Integration**: Works with existing ingress
- **Webhook Testing**: Run smoke/load tests during rollout
- **Prometheus Integration**: Uses existing metrics
- **Slack Notifications**: Alert on promotions and rollbacks

## Quick Start

### 1. Deploy Flagger

```bash
kubectl apply -k platform/progressive-delivery/flagger/
```

### 2. Create a Canary

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: my-app
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  service:
    port: 8080
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
      - name: request-success-rate
        thresholdRange:
          min: 99
```

### 3. Trigger a Deployment

```bash
kubectl set image deployment/my-app \
  my-app=ghcr.io/myorg/my-app:v2.0.0
```

Flagger will:
1. Detect the change
2. Create canary pods
3. Gradually shift traffic (10% → 20% → ... → 50%)
4. Analyze metrics at each step
5. Promote or rollback automatically

## Deployment Strategies

### Canary (Gradual Rollout)
- Traffic shifts gradually
- Metrics analyzed at each step
- Automatic rollback on failures

### A/B Testing
- Route specific users to new version
- Based on headers, cookies, etc.
- Collect metrics from both versions

### Blue/Green
- All-at-once traffic switch
- Smoke tests before switch
- Quick rollback if needed

## Metrics

Flagger uses Prometheus metrics:

```promql
# Success rate
sum(rate(http_requests_total{code!~"5.."}[1m]))
/
sum(rate(http_requests_total[1m]))

# Request duration
histogram_quantile(0.99,
  rate(http_request_duration_seconds_bucket[1m])
)
```

## Monitoring

```bash
# Watch canary progress
kubectl get canary -n production -w

# View events
kubectl describe canary my-app -n production

# Check metrics
kubectl logs -n flagger deployment/flagger
```

## Examples

- `examples/basic-canary.yaml` - Standard canary with metrics
- `examples/ab-testing.yaml` - A/B test based on headers
- `examples/blue-green.yaml` - Blue/Green with smoke tests

## Troubleshooting

### Canary Stuck

```bash
# Check flagger logs
kubectl logs -n flagger deployment/flagger

# Check metrics availability
kubectl port-forward -n monitoring svc/prometheus-server 9090
# Visit http://localhost:9090 and query metrics
```

### Metrics Not Found

Ensure your app exports Prometheus metrics:
- `/metrics` endpoint
- ServiceMonitor configured
- Prometheus scraping the service

## Resources

- [Flagger Documentation](https://docs.flagger.app/)
- [Progressive Delivery Guide](https://docs.flagger.app/usage/progressive-delivery)
