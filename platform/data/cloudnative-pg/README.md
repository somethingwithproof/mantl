# CloudNativePG - PostgreSQL Operator

Production-ready PostgreSQL with HA, backups, and point-in-time recovery.

## Features

- **High Availability**: 3-node clusters with automatic failover
- **Automated Backups**: Continuous WAL archiving + periodic base backups
- **Point-in-Time Recovery**: Restore to any point in time
- **Connection Pooling**: Built-in PgBouncer
- **Monitoring**: Prometheus metrics
- **Rolling Updates**: Zero-downtime upgrades

## Quick Start

### 1. Deploy Operator

```bash
kubectl apply -k platform/data/cloudnative-pg/operator/
```

### 2. Create PostgreSQL Cluster

```bash
kubectl apply -f platform/data/cloudnative-pg/examples/production-cluster.yaml
```

### 3. Connect to Database

```bash
# Get credentials
kubectl get secret app-db-app -n production -o jsonpath='{.data.password}' | base64 -d

# Port forward
kubectl port-forward -n production svc/app-db-rw 5432:5432

# Connect
psql -h localhost -U app -d app
```

## Backup & Recovery

### Automated Backups

Backups happen automatically:
- WAL continuous archiving
- Base backup every 6 hours (configurable)
- Retention: 30 days

### Manual Backup

```bash
kubectl cnpg backup app-db -n production
```

### Point-in-Time Recovery

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: app-db-restored
spec:
  bootstrap:
    recovery:
      source: app-db
      recoveryTarget:
        targetTime: "2026-01-24 12:00:00"
```

## Scaling

### Vertical Scaling

```bash
kubectl cnpg promote app-db-2 -n production
```

### Horizontal Scaling

```bash
# Scale to 5 replicas
kubectl patch cluster app-db -n production \
  --type merge -p '{"spec":{"instances":5}}'
```

## Monitoring

Metrics available:
- `cnpg_pg_stat_database_*`
- `cnpg_pg_replication_*`
- `cnpg_pg_locks_*`

## Connection Pooling

PgBouncer pooler:
```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Pooler
metadata:
  name: app-db-pooler
spec:
  cluster:
    name: app-db
  instances: 3
  type: rw
  pgbouncer:
    poolMode: transaction
    parameters:
      max_client_conn: "1000"
      default_pool_size: "25"
```

## Troubleshooting

```bash
# Check cluster status
kubectl cnpg status app-db -n production

# View logs
kubectl cnpg logs cluster app-db -n production

# Promote replica
kubectl cnpg promote app-db-2 -n production
```

## Resources

- [CloudNativePG Docs](https://cloudnative-pg.io/documentation/)
- [Backup Guide](https://cloudnative-pg.io/documentation/backup_recovery/)
