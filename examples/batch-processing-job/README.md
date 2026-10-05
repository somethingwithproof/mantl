# Batch Processing Job Example

Kubernetes CronJob for batch data processing demonstrating:

- CronJob scheduling for recurring batch work
- Parallel processing with worker coroutines
- Database connection pooling
- S3 integration for checkpointing
- Idempotent processing with resume capability
- Prometheus Pushgateway for job metrics
- Graceful shutdown handling
- Resource limits and security policies

## Architecture

```
┌─────────────────┐
│   CronJob       │
│  (Schedule)     │
└────────┬────────┘
         │ Triggers
         v
┌─────────────────┐
│   Job Pod       │
│  ┌───────────┐  │
│  │ Worker 1  │  │
│  │ Worker 2  │  │  ──> Process Records in Parallel
│  │ Worker 3  │  │
│  │ Worker 4  │  │
│  └───────────┘  │
└─────┬─────┬─────┘
      │     │
      │     └────────> S3 (Checkpoints)
      │
      v
┌─────────────────┐
│   PostgreSQL    │
│  (batch_items)  │
└─────────────────┘
```

## Quick Start

### Local Development

```bash
# Install dependencies
cd src
pip install -r requirements.txt

# Set environment variables
export DATABASE_URL="postgresql://user:pass@localhost:5432/db"
export S3_BUCKET="my-bucket"
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"

# Create schema and seed test data
MODE=setup python batch_processor.py

# Run batch processor
MODE=process python batch_processor.py
```

### Build Container

```bash
cd src
docker build -t batch-processor:1.0.0 .

# Run in Docker
docker run --rm \
  -e DATABASE_URL="postgresql://..." \
  -e S3_BUCKET="mantl-data" \
  -e MODE="process" \
  batch-processor:1.0.0
```

### Deploy to Kubernetes

```bash
# Update image registry
sed -i 's|REGISTRY|your-registry.example.com|g' k8s/cronjob.yaml

# Update secrets
kubectl create secret generic batch-processor-secrets \
  --from-literal=DATABASE_URL="postgresql://user:pass@host:5432/db" \
  --from-literal=AWS_ACCESS_KEY_ID="your-key" \
  --from-literal=AWS_SECRET_ACCESS_KEY="your-secret" \
  -n batch-processing \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy CronJob
kubectl apply -f k8s/cronjob.yaml

# Run setup job once
kubectl apply -f k8s/cronjob.yaml  # Creates setup job

# Verify setup completed
kubectl get jobs -n batch-processing
kubectl logs -n batch-processing job/batch-processor-setup

# Trigger CronJob manually for testing
kubectl create job --from=cronjob/batch-processor manual-run-1 -n batch-processing

# Watch job progress
kubectl get jobs -n batch-processing --watch
kubectl logs -n batch-processing -l app=batch-processor -f
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MODE` | Operation mode: `setup` or `process` | `process` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `S3_BUCKET` | S3 bucket for checkpoints | `mantl-data` |
| `BATCH_SIZE` | Records per batch | `1000` |
| `PARALLEL_WORKERS` | Number of parallel workers | `4` |
| `PUSHGATEWAY_URL` | Prometheus Pushgateway URL | `http://prometheus-pushgateway:9091` |
| `JOB_NAME` | Job identifier for metrics | `batch-processor` |

### CronJob Schedule

Current schedule: `0 2 * * *` (Daily at 2 AM UTC)

Modify the `schedule` field in `cronjob.yaml`:

```yaml
spec:
  # Every hour
  schedule: "0 * * * *"

  # Every 6 hours
  schedule: "0 */6 * * *"

  # Every Monday at 3 AM
  schedule: "0 3 * * 1"
```

## Features

### Checkpointing and Resume

The job saves its progress to S3 after each batch:

```json
{
  "last_id": 5000,
  "timestamp": "2024-03-15T10:30:00",
  "job_name": "batch-processor"
}
```

If the job fails or is interrupted, the next run will resume from the last checkpoint.

### Parallel Processing

Records are split among multiple worker coroutines:

```
Batch of 1000 records:
  Worker 0: Records 0-249
  Worker 1: Records 250-499
  Worker 2: Records 500-749
  Worker 3: Records 750-999
```

### Graceful Shutdown

The job handles SIGTERM signals gracefully:

```bash
# Send SIGTERM to pod
kubectl delete pod <pod-name> -n batch-processing

# Job will:
# 1. Complete current batch
# 2. Save checkpoint
# 3. Push metrics
# 4. Exit cleanly
```

## Monitoring

### Job Metrics

Metrics are pushed to Prometheus Pushgateway after job completion:

```promql
# Total records processed per job run
batch_records_processed_total

# Processing duration per batch
batch_processing_duration_seconds

# Job start/completion times
batch_job_start_timestamp
batch_job_completion_timestamp

# Success rate
rate(batch_records_processed_total{status="success"}[1d]) /
rate(batch_records_processed_total[1d])
```

### Job History

```bash
# List recent jobs
kubectl get jobs -n batch-processing --sort-by=.status.startTime

# View job status
kubectl describe job <job-name> -n batch-processing

# Check logs from completed job
kubectl logs -n batch-processing job/<job-name>

# Failed jobs (for debugging)
kubectl get pods -n batch-processing --field-selector=status.phase=Failed
kubectl logs -n batch-processing <failed-pod-name>
```

## Advanced Usage

### Parallel Jobs (Index-based)

For processing very large datasets, use indexed parallel jobs:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: parallel-batch-processor
spec:
  parallelism: 10  # Run 10 pods in parallel
  completions: 10  # Total number of pods to complete
  completionMode: Indexed  # Each pod gets unique index
  template:
    spec:
      containers:
      - name: processor
        image: batch-processor:1.0.0
        env:
        - name: JOB_COMPLETION_INDEX
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['batch.kubernetes.io/job-completion-index']
        # Each pod processes its assigned partition
        # Pod 0: IDs 0-999, Pod 1: IDs 1000-1999, etc.
```

### Suspending/Resuming CronJob

```bash
# Suspend CronJob (stop scheduling new runs)
kubectl patch cronjob batch-processor -n batch-processing -p '{"spec":{"suspend":true}}'

# Resume CronJob
kubectl patch cronjob batch-processor -n batch-processing -p '{"spec":{"suspend":false}}'

# Check suspension status
kubectl get cronjob batch-processor -n batch-processing -o yaml | grep suspend
```

### Manual Triggers

```bash
# Trigger job manually
kubectl create job --from=cronjob/batch-processor manual-$(date +%s) -n batch-processing

# Run with custom configuration
kubectl create job manual-large-batch -n batch-processing --dry-run=client -o yaml > job.yaml
# Edit job.yaml to modify env vars (BATCH_SIZE, etc.)
kubectl apply -f job.yaml
```

## Troubleshooting

### Job Fails Immediately

```bash
# Check job status
kubectl describe job <job-name> -n batch-processing

# Check pod events
kubectl get events -n batch-processing --field-selector involvedObject.name=<pod-name>

# Common issues:
# 1. Database connection failure
kubectl exec -it -n batch-processing <pod-name> -- \
  psql $DATABASE_URL -c "SELECT 1"

# 2. S3 access denied
kubectl logs -n batch-processing <pod-name> | grep -i "s3\|boto"

# 3. Missing secrets
kubectl get secrets -n batch-processing
```

### Job Runs Too Long

```bash
# Check current progress
kubectl logs -n batch-processing <pod-name> -f

# Check if stuck on specific batch
kubectl logs -n batch-processing <pod-name> | grep "Batch #"

# Terminate long-running job
kubectl delete job <job-name> -n batch-processing

# Adjust activeDeadlineSeconds in cronjob.yaml
```

### Out of Memory

```bash
# Check memory usage
kubectl top pod -n batch-processing

# Reduce batch size
kubectl set env cronjob/batch-processor -n batch-processing BATCH_SIZE=500

# Increase memory limits
kubectl patch cronjob batch-processor -n batch-processing --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/jobTemplate/spec/template/spec/containers/0/resources/limits/memory",
    "value": "16Gi"
  }
]'
```

## Production Considerations

1. **Idempotency**: Ensure processing logic is idempotent (can safely retry)
2. **Checkpointing**: Save progress frequently for large jobs
3. **Resource Limits**: Set appropriate CPU/memory based on data volume
4. **Concurrency Policy**: Use `Forbid` to prevent overlapping runs
5. **Cleanup**: Configure `successfulJobsHistoryLimit` to avoid clutter
6. **Monitoring**: Always push metrics to Pushgateway for completed jobs
7. **Error Handling**: Use retry logic with exponential backoff
8. **Dead Letter Queue**: Move permanently failed records to DLQ

## Cost Optimization

```bash
# Use spot instances for batch jobs (AWS)
kubectl label nodes -l node.kubernetes.io/lifecycle=spot batch-workload=true

# Add node affinity to job
kubectl patch cronjob batch-processor -n batch-processing --type=json -p='[
  {
    "op": "add",
    "path": "/spec/jobTemplate/spec/template/spec/affinity",
    "value": {
      "nodeAffinity": {
        "preferredDuringSchedulingIgnoredDuringExecution": [{
          "weight": 100,
          "preference": {
            "matchExpressions": [{
              "key": "batch-workload",
              "operator": "In",
              "values": ["true"]
            }]
          }
        }]
      }
    }
  }
]'
```

## Additional Resources

- [Kubernetes Jobs Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [CronJob Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Prometheus Pushgateway](https://github.com/prometheus/pushgateway)
