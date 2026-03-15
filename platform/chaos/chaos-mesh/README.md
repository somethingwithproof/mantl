# Chaos Mesh - Chaos Engineering

Test resilience before production incidents occur.

## Features

- **Pod Chaos**: Pod failure, kill, container kill
- **Network Chaos**: Delay, loss, duplicate, corrupt, partition
- **Stress Chaos**: CPU/memory stress
- **IO Chaos**: Delay, errno injection
- **Time Chaos**: Clock skew
- **DNS Chaos**: Random/error DNS responses
- **HTTP Chaos**: Abort, delay, replace
- **Scheduled Experiments**: Automated chaos drills

## Quick Start

### 1. Deploy Chaos Mesh

```bash
kubectl apply -k platform/chaos/chaos-mesh/operator/
```

### 2. Access Dashboard

```bash
kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333
open http://localhost:2333
```

### 3. Run Chaos Experiment

```bash
kubectl apply -f platform/chaos/chaos-mesh/experiments/pod-failure.yaml
```

## Example Experiments

### Pod Failure
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pod-kill
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - staging
    labelSelectors:
      app: api-service
  duration: "30s"
```

### Network Delay
```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-delay
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - staging
  delay:
    latency: "100ms"
  duration: "2m"
```

## GameDay Scenarios

### Scenario 1: Database Outage
```bash
# Simulate DB pod failure
kubectl apply -f - <<EOF
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: db-outage
spec:
  action: pod-failure
  mode: one
  selector:
    labelSelectors:
      app: postgres
  duration: "5m"
EOF

# Verify: Application handles gracefully
# - Circuit breaker activates
# - Queue builds up
# - Retries work
# - Alerts fire
```

### Scenario 2: High Latency
```bash
# Add network latency
kubectl apply -f - <<EOF
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: high-latency
spec:
  action: delay
  mode: all
  selector:
    namespaces:
      - production
  delay:
    latency: "500ms"
  duration: "10m"
EOF

# Verify: Timeouts configured correctly
```

### Scenario 3: Resource Exhaustion
```bash
# CPU stress
kubectl apply -f platform/chaos/chaos-mesh/experiments/stress-test.yaml

# Verify: HPA scales up
# - Alerts fire at correct thresholds
# - Performance degrades gracefully
```

## Best Practices

1. **Start in Staging**: Never test in production first
2. **Small Blast Radius**: Use `mode: one` initially
3. **Monitor Everything**: Watch metrics during experiments
4. **Document Results**: Record what worked/failed
5. **Schedule Drills**: Regular GameDays build confidence

## Monitoring Experiments

```bash
# List experiments
kubectl get podchaos,networkchaos,stresschaos -A

# Check experiment status
kubectl describe podchaos pod-failure-test

# View chaos events
kubectl get events --field-selector reason=ChaosInjected
```

## Safety

- Chaos Mesh respects namespace selectors
- Use label selectors for precise targeting
- Set duration limits
- Enable security mode
- RBAC controls who can create experiments

## Resources

- [Chaos Mesh Docs](https://chaos-mesh.org/docs/)
- [Chaos Engineering Principles](https://principlesofchaos.org/)
