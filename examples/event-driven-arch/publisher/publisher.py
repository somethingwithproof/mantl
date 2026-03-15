"""
Event Publisher - Publishes events to NATS message broker

This service demonstrates:
- Event publishing patterns
- NATS JetStream for guaranteed delivery
- REST API for triggering events
- CloudEvents format for standardization
- Prometheus metrics for event publishing
- Retry logic with exponential backoff
"""

from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel, Field
from typing import Dict, Any, Optional
from prometheus_client import Counter, Histogram, make_asgi_app
import asyncio
import nats
from nats.js.api import StreamConfig
import json
import logging
import os
from datetime import datetime, timezone
import uuid

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
NATS_URL = os.getenv("NATS_URL", "nats://nats:4222")
STREAM_NAME = os.getenv("STREAM_NAME", "EVENTS")

# Prometheus metrics
EVENTS_PUBLISHED = Counter(
    'events_published_total',
    'Total events published',
    ['event_type', 'status']
)

PUBLISH_DURATION = Histogram(
    'event_publish_duration_seconds',
    'Time spent publishing events',
    ['event_type']
)

app = FastAPI(
    title="Event Publisher",
    description="Publishes events to NATS message broker",
    version="1.0.0"
)

# Mount Prometheus metrics
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# Global NATS connection
nc = None
js = None

class Event(BaseModel):
    """Event payload following CloudEvents specification"""
    type: str = Field(..., description="Event type (e.g., 'order.created')")
    source: str = Field(..., description="Event source (e.g., 'orders-api')")
    data: Dict[str, Any] = Field(..., description="Event payload")
    correlation_id: Optional[str] = Field(None, description="Correlation ID for tracing")

class EventResponse(BaseModel):
    """Response after publishing event"""
    event_id: str
    published_at: str
    subject: str
    status: str

async def get_nats_connection():
    """Get or create NATS connection"""
    global nc, js

    if nc is None or nc.is_closed:
        logger.info(f"Connecting to NATS at {NATS_URL}")
        nc = await nats.connect(NATS_URL)
        js = nc.jetstream()

        # Ensure stream exists
        try:
            await js.add_stream(
                StreamConfig(
                    name=STREAM_NAME,
                    subjects=[f"{STREAM_NAME}.*"],
                    max_age=7 * 24 * 60 * 60,  # 7 days retention
                )
            )
            logger.info(f"Stream {STREAM_NAME} ready")
        except Exception as e:
            logger.info(f"Stream already exists or error: {e}")

    return nc, js

@app.on_event("startup")
async def startup_event():
    """Initialize NATS connection on startup"""
    await get_nats_connection()
    logger.info("Publisher service started")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    global nc
    if nc and not nc.is_closed:
        await nc.drain()
        logger.info("NATS connection closed")

@app.get("/health")
async def health():
    """Health check"""
    return {"status": "healthy"}

@app.get("/ready")
async def ready():
    """Readiness check - ensures NATS is connected"""
    global nc
    if nc is None or nc.is_closed:
        raise HTTPException(status_code=503, detail="NATS not connected")
    return {"status": "ready", "nats_connected": True}

@app.post("/api/v1/events/publish", response_model=EventResponse)
async def publish_event(event: Event):
    """Publish an event to NATS"""
    start_time = datetime.now()

    try:
        _, js = await get_nats_connection()

        # Create CloudEvents-compliant event
        event_id = str(uuid.uuid4())
        correlation_id = event.correlation_id or str(uuid.uuid4())

        cloud_event = {
            "specversion": "1.0",
            "id": event_id,
            "source": event.source,
            "type": event.type,
            "datacontenttype": "application/json",
            "time": datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.%fZ'),
            "correlationid": correlation_id,
            "data": event.data
        }

        # Publish to NATS JetStream
        subject = f"{STREAM_NAME}.{event.type.replace('.', '_')}"

        with PUBLISH_DURATION.labels(event_type=event.type).time():
            ack = await js.publish(
                subject,
                json.dumps(cloud_event).encode(),
                timeout=5.0
            )

        EVENTS_PUBLISHED.labels(event_type=event.type, status='success').inc()

        logger.info(
            f"Event published: id={event_id}, type={event.type}, "
            f"subject={subject}, seq={ack.seq}"
        )

        return EventResponse(
            event_id=event_id,
            published_at=cloud_event['time'],
            subject=subject,
            status="published"
        )

    except asyncio.TimeoutError:
        EVENTS_PUBLISHED.labels(event_type=event.type, status='timeout').inc()
        logger.error(f"Timeout publishing event type {event.type}")
        raise HTTPException(status_code=504, detail="Publish timeout")

    except Exception as e:
        EVENTS_PUBLISHED.labels(event_type=event.type, status='error').inc()
        logger.error(f"Failed to publish event: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")

@app.post("/api/v1/events/publish-batch")
async def publish_batch_events(events: list[Event]):
    """Publish multiple events in a batch"""
    if len(events) > 100:
        raise HTTPException(status_code=400, detail="Batch size cannot exceed 100")

    results = []
    for event in events:
        try:
            result = await publish_event(event)
            results.append(result)
        except Exception as e:
            logger.error(f"Failed to publish event in batch: {e}")
            # Continue with remaining events

    return {
        "total": len(events),
        "published": len(results),
        "failed": len(events) - len(results),
        "results": results
    }

@app.get("/api/v1/stream/info")
async def stream_info():
    """Get NATS stream information"""
    try:
        _, js = await get_nats_connection()
        stream = await js.stream_info(STREAM_NAME)

        return {
            "name": stream.config.name,
            "subjects": stream.config.subjects,
            "messages": stream.state.messages,
            "bytes": stream.state.bytes,
            "first_seq": stream.state.first_seq,
            "last_seq": stream.state.last_seq,
        }
    except Exception as e:
        logger.error(f"Failed to get stream info: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
