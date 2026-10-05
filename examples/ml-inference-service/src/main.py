"""
ML Inference Service - Serves machine learning model predictions via REST API

This service demonstrates:
- TensorFlow/PyTorch model serving
- Batch prediction support
- Model versioning and A/B testing
- GPU acceleration (optional)
- Prometheus metrics for model performance
- Request validation and rate limiting
"""

from fastapi import FastAPI, HTTPException, File, UploadFile, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field, validator
from typing import List, Dict, Any, Optional
from prometheus_client import Counter, Histogram, Gauge, make_asgi_app
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
import numpy as np
import logging
import time
import os
import joblib
from pathlib import Path
from urllib.parse import urlsplit

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)
MODEL_VERSION_ATTRIBUTE = "model.version"

# OpenTelemetry setup
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(__name__)


def telemetry_endpoint() -> str:
    endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://tempo:4317")
    if urlsplit(endpoint).scheme != "https" or not urlsplit(endpoint).hostname:
        raise RuntimeError("OTLP export requires an HTTPS endpoint")
    return endpoint


otlp_exporter = OTLPSpanExporter(endpoint=telemetry_endpoint(), insecure=False)
trace.get_tracer_provider().add_span_processor(BatchSpanProcessor(otlp_exporter))

# Prometheus metrics
PREDICTION_COUNT = Counter(
    "ml_predictions_total",
    "Total number of predictions made",
    ["model_name", "model_version", "status"],
)
PREDICTION_LATENCY = Histogram(
    "ml_prediction_duration_seconds",
    "Prediction latency in seconds",
    ["model_name", "model_version"],
)
BATCH_SIZE = Histogram("ml_batch_size", "Size of prediction batches", ["model_name"])
MODEL_LOAD_TIME = Gauge(
    "ml_model_load_time_seconds", "Time taken to load model", ["model_name", "model_version"]
)

app = FastAPI(
    title="ML Inference Service", description="Machine Learning Model Serving API", version="1.0.0"
)


# CORS middleware
def _parse_cors_origins() -> list[str]:
    raw = os.getenv("CORS_ALLOWED_ORIGINS", "")
    if not raw:
        logger.warning(
            "CORS_ALLOWED_ORIGINS is not set; all cross-origin requests will be rejected"
        )
        return []
    origins = [o.strip() for o in raw.split(",") if o.strip()]
    for origin in origins:
        if origin != "*" and not origin.startswith(("http://", "https://")):
            raise RuntimeError(f"Invalid CORS origin: {origin}")
    creds_raw = os.getenv("CORS_ALLOW_CREDENTIALS", "false").lower()
    if creds_raw not in ("true", "false"):
        raise RuntimeError(f"CORS_ALLOW_CREDENTIALS must be 'true' or 'false', got: {creds_raw!r}")
    allow_credentials = creds_raw == "true"
    if "*" in origins and allow_credentials:
        raise RuntimeError(
            "CORS_ALLOWED_ORIGINS=* with CORS_ALLOW_CREDENTIALS=true is unsafe; "
            "specify an explicit origin allowlist"
        )
    if "*" in origins:
        raise RuntimeError(
            "CORS_ALLOWED_ORIGINS=* is not allowed; specify explicit origin allowlist"
        )
    return origins


app.add_middleware(
    CORSMiddleware,
    allow_origins=_parse_cors_origins(),
    allow_credentials=os.getenv("CORS_ALLOW_CREDENTIALS", "false").lower() == "true",
    allow_methods=["*"],
    allow_headers=["*"],
)

# Mount Prometheus metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# Model registry
models = {}


class PredictionRequest(BaseModel):
    """Single prediction request"""

    features: Dict[str, float] = Field(..., description="Feature values for prediction")
    model_version: Optional[str] = Field("v1", description="Model version to use")

    @validator("features")
    def validate_features(cls, v):
        if not v:
            raise ValueError("Features cannot be empty")
        return v


class BatchPredictionRequest(BaseModel):
    """Batch prediction request"""

    instances: List[Dict[str, float]] = Field(..., description="List of feature dictionaries")
    model_version: Optional[str] = Field("v1", description="Model version to use")

    @validator("instances")
    def validate_instances(cls, v):
        if not v:
            raise ValueError("Instances cannot be empty")
        if len(v) > 1000:
            raise ValueError("Batch size cannot exceed 1000 instances")
        return v


class PredictionResponse(BaseModel):
    """Prediction result"""

    prediction: float
    confidence: Optional[float] = None
    model_version: str
    processing_time_ms: float


class BatchPredictionResponse(BaseModel):
    """Batch prediction results"""

    predictions: List[Dict[str, Any]]
    model_version: str
    total_processing_time_ms: float
    batch_size: int


class ModelInfo(BaseModel):
    """Model metadata"""

    name: str
    version: str
    framework: str
    input_features: List[str]
    loaded_at: str
    predictions_served: int


class DemoModel:
    """Demo ML model (simple linear regression for demonstration)"""

    def __init__(self, name: str, version: str):
        self.name = name
        self.version = version
        self.framework = "scikit-learn"
        self.input_features = ["feature1", "feature2", "feature3", "feature4"]
        self.loaded_at = time.strftime("%Y-%m-%d %H:%M:%S")
        self.predictions_served = 0

        # In production, this would load a real model:
        # self.model = joblib.load(f'/models/{name}/{version}/model.pkl')
        # For demo, we'll use simple coefficients
        self.coefficients = np.array([0.5, 0.3, 0.2, 0.1])
        self.intercept = 0.05

        logger.info(f"Loaded model {name} version {version}")

    def predict(self, features: Dict[str, float]) -> tuple:
        """Make a single prediction"""
        with tracer.start_as_current_span("model_predict") as span:
            span.set_attribute("model.name", self.name)
            span.set_attribute(MODEL_VERSION_ATTRIBUTE, self.version)

            # Convert features to array
            feature_values = [features.get(f, 0.0) for f in self.input_features]
            feature_array = np.array(feature_values)

            # Make prediction (simple linear model)
            prediction = np.dot(feature_array, self.coefficients) + self.intercept

            # Calculate confidence (simplified)
            confidence = 1.0 / (1.0 + np.abs(prediction - 0.5))

            self.predictions_served += 1

            return float(prediction), float(confidence)

    def predict_batch(self, instances: List[Dict[str, float]]) -> List[tuple]:
        """Make batch predictions"""
        with tracer.start_as_current_span("model_predict_batch") as span:
            span.set_attribute("model.name", self.name)
            span.set_attribute(MODEL_VERSION_ATTRIBUTE, self.version)
            span.set_attribute("batch.size", len(instances))

            # Convert to numpy array for batch processing
            feature_matrix = np.array(
                [[inst.get(f, 0.0) for f in self.input_features] for inst in instances]
            )

            # Batch prediction
            predictions = feature_matrix @ self.coefficients + self.intercept
            confidences = 1.0 / (1.0 + np.abs(predictions - 0.5))

            self.predictions_served += len(instances)

            return [(float(pred), float(conf)) for pred, conf in zip(predictions, confidences)]


@app.on_event("startup")
async def startup_event():
    """Load models on startup"""
    with tracer.start_as_current_span("startup"):
        logger.info("Loading models...")

        # Load multiple model versions for A/B testing
        start_time = time.time()
        models["v1"] = DemoModel("price-predictor", "v1")
        load_time = time.time() - start_time
        MODEL_LOAD_TIME.labels(model_name="price-predictor", model_version="v1").set(load_time)

        start_time = time.time()
        models["v2"] = DemoModel("price-predictor", "v2")
        load_time = time.time() - start_time
        MODEL_LOAD_TIME.labels(model_name="price-predictor", model_version="v2").set(load_time)

        logger.info(f"Loaded {len(models)} model versions")


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "models_loaded": len(models),
        "model_versions": list(models.keys()),
    }


@app.get("/ready", responses={503: {"description": "Service unavailable"}})
async def ready():
    """Readiness check - ensures models are loaded"""
    if not models:
        raise HTTPException(status_code=503, detail="Models not loaded")
    return {"status": "ready", "models": len(models)}


@app.get("/api/v1/models", response_model=List[ModelInfo])
async def list_models():
    """List available models"""
    return [
        ModelInfo(
            name=model.name,
            version=model.version,
            framework=model.framework,
            input_features=model.input_features,
            loaded_at=model.loaded_at,
            predictions_served=model.predictions_served,
        )
        for model in models.values()
    ]


@app.post(
    "/api/v1/predict",
    response_model=PredictionResponse,
    responses={
        404: {"description": "Resource not found"},
        500: {"description": "Internal operation failed"},
    },
)
async def predict(request: PredictionRequest):
    """Make a single prediction"""
    start_time = time.time()

    with tracer.start_as_current_span("predict_request") as span:
        span.set_attribute(MODEL_VERSION_ATTRIBUTE, request.model_version)

        # Get model
        model = models.get(request.model_version)
        if not model:
            PREDICTION_COUNT.labels(
                model_name="price-predictor", model_version=request.model_version, status="error"
            ).inc()
            raise HTTPException(
                status_code=404, detail=f"Model version {request.model_version} not found"
            )

        try:
            # Make prediction
            prediction, confidence = model.predict(request.features)

            processing_time = (time.time() - start_time) * 1000

            # Record metrics
            PREDICTION_COUNT.labels(
                model_name=model.name, model_version=model.version, status="success"
            ).inc()
            PREDICTION_LATENCY.labels(model_name=model.name, model_version=model.version).observe(
                time.time() - start_time
            )

            logger.info(
                f"Prediction made: model={model.version}, "
                f"prediction={prediction:.4f}, time={processing_time:.2f}ms"
            )

            return PredictionResponse(
                prediction=prediction,
                confidence=confidence,
                model_version=model.version,
                processing_time_ms=processing_time,
            )

        except Exception:
            PREDICTION_COUNT.labels(
                model_name=model.name, model_version=model.version, status="error"
            ).inc()
            logger.exception("Prediction failed")
            raise HTTPException(status_code=500, detail="Internal server error") from None


@app.post(
    "/api/v1/predict/batch",
    response_model=BatchPredictionResponse,
    responses={
        404: {"description": "Resource not found"},
        500: {"description": "Internal operation failed"},
    },
)
async def predict_batch(request: BatchPredictionRequest):
    """Make batch predictions (more efficient than individual requests)"""
    start_time = time.time()

    with tracer.start_as_current_span("predict_batch_request") as span:
        span.set_attribute(MODEL_VERSION_ATTRIBUTE, request.model_version)
        span.set_attribute("batch.size", len(request.instances))

        # Get model
        model = models.get(request.model_version)
        if not model:
            raise HTTPException(
                status_code=404, detail=f"Model version {request.model_version} not found"
            )

        try:
            # Make batch predictions
            results = model.predict_batch(request.instances)

            processing_time = (time.time() - start_time) * 1000

            # Format results
            predictions = [
                {"prediction": pred, "confidence": conf, "instance_index": i}
                for i, (pred, conf) in enumerate(results)
            ]

            # Record metrics
            PREDICTION_COUNT.labels(
                model_name=model.name, model_version=model.version, status="success"
            ).inc(len(request.instances))
            BATCH_SIZE.labels(model_name=model.name).observe(len(request.instances))
            PREDICTION_LATENCY.labels(model_name=model.name, model_version=model.version).observe(
                time.time() - start_time
            )

            logger.info(
                f"Batch prediction complete: model={model.version}, "
                f"batch_size={len(request.instances)}, time={processing_time:.2f}ms"
            )

            return BatchPredictionResponse(
                predictions=predictions,
                model_version=model.version,
                total_processing_time_ms=processing_time,
                batch_size=len(request.instances),
            )

        except Exception:
            PREDICTION_COUNT.labels(
                model_name=model.name, model_version=model.version, status="error"
            ).inc()
            logger.exception("Batch prediction failed")
            raise HTTPException(status_code=500, detail="Internal server error") from None


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host=os.getenv("HOST", "127.0.0.1"), port=8000)
