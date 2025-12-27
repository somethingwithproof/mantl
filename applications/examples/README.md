# Mantl Example Applications

This directory contains example applications demonstrating various platform features and best practices.

## Available Examples

### 1. Hello World (`hello/`)
Simple deployment demonstrating basic Kubernetes workload deployment.

**Features:**
- Basic Deployment + Service
- ArgoCD GitOps deployment
- Minimal resource requirements

**Deploy:**
```bash
kubectl apply -k applications/examples/hello
```

### 2. API Service (`api-service/`)
Backend API service demonstrating observability and security best practices.

**Features:**
- OpenTelemetry distributed tracing integration
- Prometheus metrics via ServiceMonitor
- Security context (non-root, read-only filesystem)
- Health checks (liveness + readiness)
- Resource limits and requests

**Deploy:**
```bash
kubectl apply -k applications/examples/api-service/base
```

**Verify:**
```bash
# Check pods
kubectl get pods -l app.kubernetes.io/name=api-service

# View metrics
kubectl port-forward svc/api-service 9090:9090
curl http://localhost:9090/metrics

# Check traces in Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Open http://localhost:3000, navigate to Explore → Tempo
```

### 3. Frontend (`frontend/`)
Web frontend demonstrating Gateway API ingress and multi-tier architecture.

**Features:**
- Nginx-based frontend
- Gateway API HTTPRoute for ingress
- ConfigMap-based configuration
- Security context (non-root, read-only filesystem)
- Backend API proxy configuration

**Deploy:**
```bash
kubectl apply -k applications/examples/frontend/base
```

**Access:**
```bash
# Via Gateway API (requires DNS or /etc/hosts entry)
curl http://frontend.example.com

# Or via port-forward
kubectl port-forward svc/frontend 8080:80
curl http://localhost:8080
```

### 4. Database Application (`database-app/`)
Stateful application demonstrating persistent storage and database integration.

**Features:**
- PostgreSQL StatefulSet with persistent volume
- Application Deployment with database connectivity
- Init container for database readiness
- Secret management for credentials
- Headless service for StatefulSet

**Deploy:**
```bash
kubectl apply -k applications/examples/database-app/base
```

**Verify:**
```bash
# Check StatefulSet
kubectl get statefulset postgres
kubectl get pvc -l app.kubernetes.io/name=postgres

# Check database connectivity
kubectl exec -it postgres-0 -- psql -U appuser -d appdb -c "SELECT version();"

# Check app logs
kubectl logs -l app.kubernetes.io/name=database-app
```

## Best Practices Demonstrated

### Security
- ✅ Non-root containers (`runAsNonRoot: true`)
- ✅ Read-only root filesystem where possible
- ✅ Dropped all capabilities (`capabilities.drop: [ALL]`)
- ✅ SeccompProfile set to `RuntimeDefault`
- ✅ ServiceAccount per application
- ✅ No privilege escalation

### Observability
- ✅ Prometheus metrics via ServiceMonitor (api-service)
- ✅ OpenTelemetry tracing integration (api-service)
- ✅ Health checks (liveness + readiness)
- ✅ Structured logging

### GitOps
- ✅ Kustomize-based configuration
- ✅ Environment-specific overlays (production/staging/dev)
- ✅ ArgoCD Application manifests
- ✅ Proper labeling strategy

### Resource Management
- ✅ CPU and memory requests defined
- ✅ CPU and memory limits set
- ✅ Appropriate replica counts
- ✅ Horizontal Pod Autoscaler ready (can be added)

## Production Considerations

Before deploying to production:

1. **Replace Example Images**: Update image references in deployments
2. **Configure Secrets**: Use External Secrets Operator instead of plaintext secrets
3. **Set Resource Limits**: Adjust based on actual workload requirements
4. **Enable Autoscaling**: Add HorizontalPodAutoscaler for dynamic scaling
5. **Configure Ingress**: Update HTTPRoute hostnames to your domain
6. **Enable TLS**: Add cert-manager annotations for automatic TLS
7. **Add Monitoring**: Create Grafana dashboards for your applications
8. **Set Backup**: Configure database backups for stateful applications

## ArgoCD Deployment

Create ArgoCD Applications for each example:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: api-service
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/thomasvincent/mantl.git
    targetRevision: HEAD
    path: applications/examples/api-service/base
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## Cleanup

Remove all example applications:

```bash
kubectl delete -k applications/examples/api-service/base
kubectl delete -k applications/examples/frontend/base
kubectl delete -k applications/examples/database-app/base
kubectl delete -k applications/examples/hello
```
