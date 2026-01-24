# Velero - Backup & Disaster Recovery

Automated cluster and namespace-level backup/restore.

## Features

- **Automated Backups**: Daily and weekly schedules
- **Point-in-Time Recovery**: Restore to any backup
- **Volume Snapshots**: CSI snapshot integration
- **Cross-Region Replication**: Multi-region DR
- **Namespace-Level Backups**: Selective backup/restore

## Quick Start

### 1. Configure Cloud Storage

**AWS S3:**
```bash
# Create S3 bucket
aws s3 mb s3://mantl-velero-backups --region us-east-1

# Create IAM role with IRSA
# See: docs/velero-aws-setup.md
```

**GCP:**
```bash
# Create GCS bucket
gsutil mb gs://mantl-velero-backups

# Configure Workload Identity
# See: docs/velero-gcp-setup.md
```

**Azure:**
```bash
# Create storage account
az storage account create \
  --name mantlvelerobackups \
  --resource-group mantl-backups

# See: docs/velero-azure-setup.md
```

### 2. Deploy Velero

```bash
kubectl apply -k platform/backup/velero/
```

### 3. Create On-Demand Backup

```bash
# Backup entire cluster
velero backup create full-backup

# Backup specific namespace
velero backup create production-backup --include-namespaces production

# Backup with volume snapshots
velero backup create app-backup --include-namespaces production --snapshot-volumes
```

### 4. Restore from Backup

```bash
# List backups
velero backup get

# Restore
velero restore create --from-backup production-backup

# Restore to different namespace
velero restore create --from-backup production-backup \
  --namespace-mappings production:production-restore
```

## Backup Schedules

### Daily Backup (2 AM)
- Namespaces: production, staging
- Retention: 30 days
- Volume snapshots: Yes

### Weekly Backup (Sunday 3 AM)
- Namespaces: All
- Retention: 90 days
- Volume snapshots: Yes

## DR Procedures

### Scenario 1: Namespace Corruption

```bash
# Delete corrupted namespace
kubectl delete namespace production

# Restore from last backup
velero restore create production-restore \
  --from-backup daily-backup-20260124
```

### Scenario 2: Cluster Failure

```bash
# In new cluster:
1. Install Velero pointing to same bucket
2. velero backup get
3. velero restore create --from-backup full-backup
```

### Scenario 3: Accidental Deletion

```bash
# Restore specific resources
velero restore create --from-backup production-backup \
  --include-resources deployments,services \
  --selector app=critical-app
```

## Monitoring

```bash
# Check backup status
velero backup describe daily-backup-20260124

# Check restore status
velero restore describe production-restore

# View logs
velero backup logs daily-backup-20260124
```

## Backup Validation

Automated restore testing:
```bash
# Test restore to temporary namespace
velero restore create test-restore \
  --from-backup daily-backup-20260124 \
  --namespace-mappings production:test-restore

# Verify
kubectl get all -n test-restore

# Cleanup
kubectl delete namespace test-restore
```

## Troubleshooting

### Backup Fails

```bash
velero backup describe BACKUP_NAME
velero backup logs BACKUP_NAME
```

### Volume Snapshot Issues

Check CSI driver:
```bash
kubectl get volumesnapshotclass
kubectl get volumesnapshot -A
```

## Resources

- [Velero Documentation](https://velero.io/docs/)
- [Disaster Recovery Guide](https://velero.io/docs/disaster-case/)
