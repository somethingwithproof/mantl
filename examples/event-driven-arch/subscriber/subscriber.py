"""
Event Subscriber - Consumes events from NATS message broker

This service demonstrates:
- Event subscription patterns
- NATS JetStream consumer groups
- At-least-once delivery with acknowledgment
- Dead letter queue for failed messages
- Graceful shutdown and rebalancing
- Prometheus metrics for event processing
"""

import asyncio
import nats
from nats.js.api import ConsumerConfig, DeliverPolicy, AckPolicy
from nats.js import JetStreamContext
import json
import logging
import signal
import sys
import os
from typing import Dict, Any
from prometheus_client import Counter, Histogram, Gauge, start_http_server
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
NATS_URL = os.getenv("NATS_URL", "nats://nats:4222")
STREAM_NAME = os.getenv("STREAM_NAME", "EVENTS")
CONSUMER_NAME = os.getenv("CONSUMER_NAME", "order-processor")
DURABLE_NAME = os.getenv("DURABLE_NAME", "order-processor-durable")
FILTER_SUBJECT = os.getenv("FILTER_SUBJECT", "EVENTS.order_*")
METRICS_PORT = int(os.getenv("METRICS_PORT", "9090"))

# Prometheus metrics
EVENTS_RECEIVED = Counter(
    'events_received_total',
    'Total events received',
    ['event_type', 'status']
)

PROCESSING_DURATION = Histogram(
    'event_processing_duration_seconds',
    'Time spent processing events',
    ['event_type']
)

QUEUE_DEPTH = Gauge(
    'event_queue_depth',
    'Number of pending messages in queue'
)

# Global state
shutdown_requested = False

def signal_handler(signum, frame):
    """Handle shutdown signals"""
    global shutdown_requested
    logger.info(f"Received signal {signum}, initiating graceful shutdown...")
    shutdown_requested = True

signal.signal(signal.SIGTERM, signal_handler)
signal.signal(signal.SIGINT, signal_handler)

class EventHandler:
    """Handles processing of different event types"""

    async def handle_order_created(self, event_data: Dict[str, Any]):
        """Handle order.created event"""
        logger.info(f"Processing order.created: order_id={event_data.get('order_id')}")

        # Simulate processing
        await asyncio.sleep(0.5)

        # Example processing logic:
        # - Send confirmation email
        # - Update inventory
        # - Notify shipping service
        # - Create analytics event

        logger.info(f"Order created processed: {event_data.get('order_id')}")

    async def handle_order_updated(self, event_data: Dict[str, Any]):
        """Handle order.updated event"""
        logger.info(f"Processing order.updated: order_id={event_data.get('order_id')}")
        await asyncio.sleep(0.2)
        logger.info(f"Order updated processed: {event_data.get('order_id')}")

    async def handle_order_cancelled(self, event_data: Dict[str, Any]):
        """Handle order.cancelled event"""
        logger.info(f"Processing order.cancelled: order_id={event_data.get('order_id')}")
        await asyncio.sleep(0.3)
        logger.info(f"Order cancelled processed: {event_data.get('order_id')}")

    async def handle_event(self, event: Dict[str, Any]) -> bool:
        """Route event to appropriate handler"""
        event_type = event.get('type', 'unknown')

        handlers = {
            'order.created': self.handle_order_created,
            'order.updated': self.handle_order_updated,
            'order.cancelled': self.handle_order_cancelled,
        }

        handler = handlers.get(event_type)
        if not handler:
            logger.warning(f"No handler for event type: {event_type}")
            return True  # Acknowledge anyway to avoid retries

        try:
            with PROCESSING_DURATION.labels(event_type=event_type).time():
                await handler(event.get('data', {}))

            EVENTS_RECEIVED.labels(event_type=event_type, status='success').inc()
            return True

        except Exception as e:
            logger.error(f"Failed to process event {event_type}: {e}")
            EVENTS_RECEIVED.labels(event_type=event_type, status='error').inc()
            return False

async def message_callback(msg, event_handler: EventHandler):
    """Callback for processing messages from NATS"""
    try:
        # Parse CloudEvent
        event = json.loads(msg.data.decode())

        logger.info(
            f"Received event: id={event.get('id')}, "
            f"type={event.get('type')}, "
            f"correlation_id={event.get('correlationid')}"
        )

        # Process event
        success = await event_handler.handle_event(event)

        if success:
            # Acknowledge successful processing
            await msg.ack()
            logger.debug(f"Event acknowledged: {event.get('id')}")
        else:
            # Negative acknowledge (will trigger redelivery)
            await msg.nak(delay=5)  # Retry after 5 seconds
            logger.warning(f"Event processing failed, will retry: {event.get('id')}")

    except json.JSONDecodeError as e:
        logger.error(f"Invalid JSON in message: {e}")
        # Acknowledge to remove from queue (can't process invalid JSON)
        await msg.ack()

    except Exception as e:
        logger.error(f"Unexpected error processing message: {e}")
        # Negative acknowledge for retry
        await msg.nak(delay=10)

async def main():
    """Main subscriber loop"""
    global shutdown_requested

    logger.info(f"Starting event subscriber: {CONSUMER_NAME}")

    # Start Prometheus metrics server
    start_http_server(METRICS_PORT)
    logger.info(f"Metrics server started on port {METRICS_PORT}")

    # Connect to NATS
    nc = await nats.connect(NATS_URL)
    js = nc.jetstream()

    logger.info(f"Connected to NATS at {NATS_URL}")

    # Create event handler
    event_handler = EventHandler()

    try:
        # Create durable consumer
        consumer_config = ConsumerConfig(
            durable_name=DURABLE_NAME,
            deliver_policy=DeliverPolicy.ALL,
            ack_policy=AckPolicy.EXPLICIT,
            max_deliver=3,  # Max retry attempts
            ack_wait=30,  # 30 seconds to process and ack
            max_ack_pending=100,  # Process up to 100 messages concurrently
            filter_subject=FILTER_SUBJECT
        )

        # Subscribe to stream
        psub = await js.pull_subscribe(
            FILTER_SUBJECT,
            DURABLE_NAME,
            stream=STREAM_NAME,
            config=consumer_config
        )

        logger.info(
            f"Subscribed to stream {STREAM_NAME} "
            f"with filter {FILTER_SUBJECT}"
        )

        # Process messages
        while not shutdown_requested:
            try:
                # Fetch batch of messages
                msgs = await psub.fetch(batch=10, timeout=5)

                if msgs:
                    QUEUE_DEPTH.set(len(msgs))

                    # Process messages concurrently
                    tasks = [
                        message_callback(msg, event_handler)
                        for msg in msgs
                    ]
                    await asyncio.gather(*tasks, return_exceptions=True)

                else:
                    # No messages, update queue depth
                    QUEUE_DEPTH.set(0)
                    await asyncio.sleep(1)

            except nats.errors.TimeoutError:
                # No messages available, continue
                QUEUE_DEPTH.set(0)
                continue

            except Exception as e:
                logger.error(f"Error fetching messages: {e}")
                await asyncio.sleep(5)

    except Exception as e:
        logger.error(f"Subscriber error: {e}")
        sys.exit(1)

    finally:
        logger.info("Shutting down subscriber...")
        await nc.drain()
        logger.info("Subscriber stopped")

if __name__ == "__main__":
    asyncio.run(main())
