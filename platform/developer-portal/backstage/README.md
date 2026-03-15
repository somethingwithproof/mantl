# Backstage Developer Portal

Self-service platform with service catalog and golden paths.

## Features

- **Service Catalog**: All services, APIs, and resources in one place
- **Scaffolding Templates**: Create new services with golden paths
- **Tech Docs**: Documentation integrated with services
- **Kubernetes Plugin**: View pod status, logs from portal
- **Cost Tracking**: Per-service cost visibility (via OpenCost)
- **Search**: Find any service, API, or resource

## Quick Start

### 1. Deploy Backstage

```bash
kubectl apply -k platform/developer-portal/backstage/
```

### 2. Access Portal

```bash
kubectl port-forward -n backstage svc/backstage 7007:7007
open http://localhost:7007
```

### 3. Register Your First Service

Create `catalog-info.yaml` in your service repo:
```yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: api-service
  description: Main API service
  annotations:
    github.com/project-slug: thomasvincent/api-service
    backstage.io/kubernetes-id: api-service
spec:
  type: service
  lifecycle: production
  owner: team-platform
  system: core
```

## Creating New Services

### From Template

1. Go to **Create** → **Choose Template**
2. Select "Microservice"
3. Fill in details:
   - Name
   - Description
   - Owner team
   - Runtime (Node.js, Python, Go, Java)
4. Click **Create**

Backstage will:
- Create GitHub repository
- Add CI/CD pipeline
- Deploy to Kubernetes
- Register in service catalog

## Service Catalog

### Viewing Services

```bash
# All services
https://backstage.mantl.example.com/catalog

# By owner
https://backstage.mantl.example.com/catalog?filters[owner]=team-platform

# By type
https://backstage.mantl.example.com/catalog?filters[kind]=API
```

### Service Details

Each service shows:
- Owner and lifecycle
- Dependencies (APIs, databases)
- Kubernetes status
- Recent deployments
- Cost allocation
- Documentation

## Kubernetes Plugin

View from portal:
- Pod status and count
- Recent logs
- Resource usage
- Errors and crashes

## Tech Docs

Write docs in Markdown:
```markdown
# docs/index.md

# API Service Documentation

## Getting Started
...
```

Add to `catalog-info.yaml`:
```yaml
metadata:
  annotations:
    backstage.io/techdocs-ref: dir:.
```

## Cost Visibility

OpenCost integration shows:
- Monthly cost per service
- Cost trend (30 days)
- Most expensive resources

## Search

Press `/` to search:
- Services by name
- APIs by endpoint
- Resources by type
- Documentation

## Resources

- [Backstage Docs](https://backstage.io/docs)
- [Template Guide](https://backstage.io/docs/features/software-templates/)
