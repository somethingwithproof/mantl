<!-- SPDX-License-Identifier: Apache-2.0 -->
# ML Inference Service Example

Machine learning model serving application demonstrating:

- FastAPI-based REST API for model predictions
- Single and batch prediction endpoints
- Model versioning and A/B testing
- Prometheus metrics for model performance monitoring
- OpenTelemetry distributed tracing
- Horizontal autoscaling based on latency
- Optional GPU acceleration support
- Production-ready security and reliability patterns

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       v
┌─────────────────────┐
│  Ingress (NGINX)    │
│  - Rate limiting    │
│  - TLS termination  │
└──────┬──────────────┘
       │
       v
┌─────────────────────┐
│  ML Inference API   │
│  - FastAPI          │
│  - Model v1, v2     │
│  - Batch support    │
└──────┬──────────────┘
       │
       v
┌─────────────────────┐
│  Shared Storage     │
│  - EFS/NFS          │
│  - Model files      │
└─────────────────────┘
```

## Quick Start

### Local Development

```bash
# Install dependencies
cd src
pip install -r requirements.txt

# Run locally
python main.py

# Test prediction
curl -X POST http://localhost:8000/api/v1/predict \
  -H "Content-Type: application/json" \
  -d '{
    "features": {
      "feature1": 1.0,
      "feature2": 2.0,
      "feature3": 3.0,
      "feature4": 4.0
    },
    "model_version": "v1"
  }'
```

### Build Container

```bash
cd src
docker build -t ml-inference:1.0.0 .

# Run container
docker run -p 8000:8000 ml-inference:1.0.0
```

### Deploy to Kubernetes

```bash
# Update image registry in deployment.yaml
sed -i 's|REGISTRY|your-registry.example.com|g' k8s/deployment.yaml

# Deploy
kubectl apply -f k8s/deployment.yaml

# Verify deployment
kubectl get pods -n ml-inference
kubectl logs -n ml-inference -l app=ml-inference

# Test via port-forward
kubectl port-forward -n ml-inference svc/ml-inference 8000:8000

curl -X POST http://localhost:8000/api/v1/predict \
  -H "Content-Type: application/json" \
  -d '{
    "features": {
      "feature1": 1.5,
      "feature2": 2.5,
      "feature3": 3.5,
      "feature4": 4.5
    }
  }'
```

## API Endpoints

### Single Prediction

```bash
POST /api/v1/predict

Request:
{
  "features": {
    "feature1": 1.0,
    "feature2": 2.0,
    "feature3": 3.0,
    "feature4": 4.0
  },
  "model_version": "v1"  // optional, defaults to "v1"
}

Response:
{
  "prediction": 0.85,
  "confidence": 0.92,
  "model_version": "v1",
  "processing_time_ms": 12.5
}
```

### Batch Prediction

```bash
POST /api/v1/predict/batch

Request:
{
  "instances": [
    {"feature1": 1.0, "feature2": 2.0, "feature3": 3.0, "feature4": 4.0},
    {"feature1": 1.5, "feature2": 2.5, "feature3": 3.5, "feature4": 4.5},
    {"feature1": 2.0, "feature2": 3.0, "feature3": 4.0, "feature4": 5.0}
  ],
  "model_version": "v1"
}

Response:
{
  "predictions": [
    {"prediction": 0.85, "confidence": 0.92, "instance_index": 0},
    {"prediction": 0.90, "confidence": 0.88, "instance_index": 1},
    {"prediction": 0.95, "confidence": 0.85, "instance_index": 2}
  ],
  "model_version": "v1",
  "total_processing_time_ms": 15.3,
  "batch_size": 3
}
```

### List Models

```bash
GET /api/v1/models

Response:
[
  {
    "name": "price-predictor",
    "version": "v1",
    "framework": "scikit-learn",
    "input_features": ["feature1", "feature2", "feature3", "feature4"],
    "loaded_at": "2024-03-15 10:30:00",
    "predictions_served": 1234
  }
]
```

## Model Management

### Loading Custom Models

Replace the `DemoModel` class with actual model loading:

```python
import joblib
import tensorflow as tf
import torch

class ProductionModel:
    def __init__(self, model_path: str):
        # For scikit-learn
        self.model = joblib.load(model_path)

        # For TensorFlow
        # self.model = tf.keras.models.load_model(model_path)

        # For PyTorch
        # self.model = torch.load(model_path)
        # self.model.set_to_inference_mode()

    def predict(self, features):
        return self.model.predict([features])[0]
```

### Model Updates (Zero Downtime)

```bash
# 1. Upload new model to shared storage
kubectl cp new-model.pkl ml-inference/ml-inference-0:/app/models/v3/

# 2. Update deployment to load new version
kubectl set env deployment/ml-inference -n ml-inference \
  MODEL_VERSION_V3_ENABLED=true

# 3. Monitor new model performance
kubectl logs -n ml-inference -l version=v2 -f | grep prediction

# 4. Gradually increase traffic to new version
# 5. Deprecate old version after validation
```

## Monitoring

### Key Metrics

```promql
# Prediction latency (95th percentile)
histogram_quantile(0.95, rate(ml_prediction_duration_seconds_bucket[5m]))

# Prediction throughput
rate(ml_predictions_total[5m])

# Error rate
rate(ml_predictions_total{status="error"}[5m]) / rate(ml_predictions_total[5m])

# Model usage distribution
sum by (model_version) (rate(ml_predictions_total[5m]))

# Batch size distribution
histogram_quantile(0.90, rate(ml_batch_size_bucket[5m]))
```

## GPU Support

### Enable GPU Nodes

```bash
# Create GPU node pool (GKE example)
gcloud container node-pools create gpu-pool \
  --cluster=mantl-prod \
  --accelerator=type=nvidia-tesla-t4,count=1 \
  --machine-type=n1-standard-4 \
  --num-nodes=1 \
  --min-nodes=0 \
  --max-nodes=5

# Install NVIDIA device plugin
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.14.0/nvidia-device-plugin.yml
```

### Update Deployment for GPU

Uncomment GPU sections in `deployment.yaml`:
- `resources.limits.nvidia.com/gpu`
- `nodeSelector.accelerator`
- `tolerations` for GPU nodes

## Performance Optimization

### Batch Predictions

Always prefer batch endpoint for multiple predictions:

```python
# Efficient: Single batch request
response = requests.post("/api/v1/predict/batch", json={
    "instances": [instance1, instance2, instance3]
})

# Inefficient: Multiple individual requests
for instance in instances:
    response = requests.post("/api/v1/predict", json={"features": instance})
```

## Production Considerations

1. **Model Versioning**: Use semantic versioning, tag images with model version
2. **Model Monitoring**: Track prediction drift and accuracy degradation
3. **Resource Planning**: GPU instances are expensive, use autoscaling
4. **Rate Limiting**: Protect from abuse, enforce quotas per client
5. **Security**: Validate input, sanitize outputs, use authentication

## Additional Resources

- [TensorFlow Serving](https://www.tensorflow.org/tfx/guide/serving)
- [TorchServe](https://pytorch.org/serve/)
- [KServe Documentation](https://kserve.github.io/website/)
