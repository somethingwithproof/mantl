<!-- SPDX-License-Identifier: Apache-2.0 -->
# Event-Driven Architecture Example

Microservices communication using NATS message broker demonstrating:

- Publisher/subscriber pattern with NATS JetStream
- CloudEvents standard for event formatting
- At-least-once delivery guarantees
- Consumer groups for load distribution
- Dead letter queue for failed messages
- Graceful shutdown and message acknowledgment
- Prometheus metrics for event tracking
- Horizontal autoscaling based on queue depth

## Architecture

```
┌──────────────┐      ┌─────────────────┐      ┌──────────────┐
│   Producer   │─────>│  NATS JetStream │─────>│  Subscriber  │
│   Service    │      │   (Clustered)   │      │  (Replicas)  │
└──────────────┘      └─────────────────┘      └──────────────┘
                              │
                              v
                      ┌───────────────┐
                      │  Persistent   │
                      │    Storage    │
                      └───────────────┘
```

### Components

1. **Publisher Service**: REST API that publishes events to NATS
2. **Subscriber Service**: Long-running service that consumes and processes events
3. **NATS Cluster**: 3-node JetStream cluster for message persistence

## Quick Start

### Local Development

```bash
# Start NATS locally
docker run -p 4222:4222 nats:2.10-alpine -js

# Run publisher
cd publisher
pip install -r requirements.txt
python publisher.py

# In another terminal, run subscriber
cd subscriber
pip install -r requirements.txt
python subscriber.py

# Publish a test event
curl -X POST http://localhost:8000/api/v1/events/publish \
  -H "Content-Type: application/json" \
  -d '{
    "type": "order.created",
    "source": "orders-api",
    "data": {
      "order_id": "12345",
      "customer_id": "67890",
      "total": 99.99
    }
  }'
```

### Deploy to Kubernetes

```bash
# Build and push these versioned example images to your registry first
# publisher:1.0.0 and subscriber:1.0.0
# Use immutable image digests for a production overlay.
# Update image registries
sed -i 's|REGISTRY|your-registry.example.com|g' k8s/*.yaml

# Deploy NATS cluster
kubectl apply -f k8s/nats.yaml

# Wait for NATS to be ready
kubectl wait --for=condition=Ready pod -l app=nats -n event-driven --timeout=5m

# Deploy publisher
kubectl apply -f k8s/publisher.yaml

# Deploy subscriber
kubectl apply -f k8s/subscriber.yaml

# Verify deployments
kubectl get pods -n event-driven
kubectl logs -n event-driven -l app=event-subscriber -f
```

## Event Publishing

### Publish Single Event

```bash
curl -X POST http://events.example.com/api/v1/events/publish \
  -H "Content-Type: application/json" \
  -d '{
    "type": "order.created",
    "source": "orders-api",
    "data": {
      "order_id": "12345",
      "customer_id": "67890",
      "items": [
        {"product_id": "A1", "quantity": 2, "price": 29.99}
      ],
      "total": 99.99,
      "status": "pending"
    },
    "correlation_id": "request-abc-123"
  }'

Response:
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "published_at": "2024-03-15T10:30:00Z",
  "subject": "EVENTS.order_created",
  "status": "published"
}
```

### Publish Batch Events

```bash
curl -X POST http://events.example.com/api/v1/events/publish-batch \
  -H "Content-Type: application/json" \
  -d '[
    {
      "type": "order.created",
      "source": "orders-api",
      "data": {"order_id": "001"}
    },
    {
      "type": "order.updated",
      "source": "orders-api",
      "data": {"order_id": "002"}
    }
  ]'
```

## Event Types

The example implements handlers for:

| Event Type | Description | Subscriber Action |
|------------|-------------|-------------------|
| `order.created` | New order placed | Send confirmation, update inventory |
| `order.updated` | Order modified | Update records, notify customer |
| `order.cancelled` | Order cancelled | Refund, restore inventory |

### Adding New Event Types

1. Add handler in `subscriber.py`:
```python
async def handle_payment_completed(self, event_data: Dict[str, Any]):
    logger.info(f"Processing payment: {event_data.get('payment_id')}")
    # Your processing logic
```

2. Register in handlers dict:
```python
handlers = {
    'order.created': self.handle_order_created,
    'payment.completed': self.handle_payment_completed,  # New
}
```

3. Update FILTER_SUBJECT to include new event:
```yaml
env:
- name: FILTER_SUBJECT
  value: "EVENTS.order_*,EVENTS.payment_*"
```

## Monitoring

### NATS Cluster Health

```bash
# Check NATS cluster status
kubectl exec -it nats-0 -n event-driven -- nats server list

# Check stream info
kubectl exec -it nats-0 -n event-driven -- nats stream info EVENTS

# Check consumer info
kubectl exec -it nats-0 -n event-driven -- nats consumer info EVENTS order-processor-durable
```

### Event Metrics

```promql
# Publishing rate
rate(events_published_total[5m])

# Publishing success rate
rate(events_published_total{status="success"}[5m]) /
rate(events_published_total[5m])

# Processing rate
rate(events_received_total[5m])

# Queue depth (backlog)
event_queue_depth

# Processing latency
histogram_quantile(0.95, rate(event_processing_duration_seconds_bucket[5m]))
```

### Subscriber Logs

```bash
# Watch subscriber logs
kubectl logs -n event-driven -l app=event-subscriber -f

# Check for errors
kubectl logs -n event-driven -l app=event-subscriber --tail=100 | grep ERROR

# View specific event processing
kubectl logs -n event-driven -l app=event-subscriber | grep "event_id=<id>"
```

## Scaling

### Scale Publishers (Stateless)

```bash
# Manual scaling
kubectl scale deployment event-publisher -n event-driven --replicas=5

# Auto-scaling based on request rate
kubectl autoscale deployment event-publisher \
  --namespace event-driven \
  --cpu-percent=70 \
  --min=2 \
  --max=10
```

### Scale Subscribers (Consumer Group)

```bash
# Add more subscribers to process events faster
kubectl scale deployment event-subscriber -n event-driven --replicas=10

# NATS will automatically distribute messages among replicas
# Each message is delivered to exactly one subscriber (consumer group pattern)
```

### Scale NATS Cluster

```bash
# Add more NATS nodes for higher throughput
kubectl scale statefulset nats -n event-driven --replicas=5

# Update cluster routes in ConfigMap
kubectl edit configmap nats-config -n event-driven
```

## Troubleshooting

### Events Not Being Delivered

```bash
# Check publisher connectivity
kubectl exec -it -n event-driven deploy/event-publisher -- \
  nc -zv nats 4222

# Check stream exists
kubectl exec -it nats-0 -n event-driven -- nats stream list

# Check consumer exists
kubectl exec -it nats-0 -n event-driven -- nats consumer list EVENTS

# Check for errors in publisher
kubectl logs -n event-driven -l app=event-publisher | grep ERROR
```

### Subscriber Not Processing

```bash
# Check subscriber logs
kubectl logs -n event-driven -l app=event-subscriber -f

# Check if subscriber is connected
kubectl logs -n event-driven -l app=event-subscriber | grep "Connected to NATS"

# Check consumer lag
kubectl exec -it nats-0 -n event-driven -- \
  nats consumer info EVENTS order-processor-durable
```

### Message Backlog Growing

```bash
# Check queue depth
kubectl exec -it nats-0 -n event-driven -- \
  nats stream info EVENTS | grep "Messages:"

# Scale up subscribers
kubectl scale deployment event-subscriber -n event-driven --replicas=10

# Check processing rate
kubectl logs -n event-driven -l app=event-subscriber | grep "Received event"
```

## Advanced Patterns

### Dead Letter Queue

For messages that fail after max retries:

```bash
# Configure DLQ stream
kubectl exec -it nats-0 -n event-driven -- nats stream add DLQ \
  --subjects "DLQ.>" \
  --retention limits \
  --max-age 30d

# Update consumer to send to DLQ after max retries
# (Requires code modification to check delivery count and publish to DLQ)
```

### Event Replay

```bash
# Replay all events from stream
kubectl exec -it nats-0 -n event-driven -- \
  nats consumer add EVENTS replay \
  --filter EVENTS.order_created \
  --deliver all \
  --max-deliver 1

# Create temporary subscriber to process replayed events
```

### Priority Queues

Create separate streams for different priorities:

```yaml
# High priority stream
- name: EVENTS_HIGH_PRIORITY
  subjects: ["EVENTS.high.*"]
  max_age: 1h

# Normal priority stream
- name: EVENTS
  subjects: ["EVENTS.*"]
  max_age: 7d
```

## Production Considerations

1. **Event Schema**: Define and version event schemas (JSON Schema, Avro, Protobuf)
2. **Idempotency**: Ensure event handlers are idempotent (can process same event multiple times)
3. **Ordering**: NATS preserves order within a subject, not across subjects
4. **Retention**: Configure appropriate message retention based on replay requirements
5. **Monitoring**: Alert on queue depth, processing lag, error rates
6. **Testing**: Test failure scenarios (subscriber crash, NATS unavailable, slow processing)
7. **Dead Letter Queue**: Implement DLQ for messages that fail repeatedly
8. **Correlation IDs**: Use for distributed tracing across services

## Additional Resources

- [NATS Documentation](https://docs.nats.io/)
- [CloudEvents Specification](https://cloudevents.io/)
- [Event-Driven Architecture Patterns](https://www.enterpriseintegrationpatterns.com/)

Build and publish the versioned images before applying these templates:

```sh
docker build -t YOUR_REGISTRY/event-publisher:1.0.0 publisher/
docker build -t YOUR_REGISTRY/event-subscriber:1.0.0 subscriber/
docker push YOUR_REGISTRY/event-publisher:1.0.0
docker push YOUR_REGISTRY/event-subscriber:1.0.0
```

Replace `YOUR_REGISTRY` consistently in the manifests. Container scratch-space
requests and limits bound local writable data separately from NATS’s persistent
volume; tune both from measurements of your own workload.
