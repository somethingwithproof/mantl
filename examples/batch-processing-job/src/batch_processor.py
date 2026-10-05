"""
Batch Processing Job - Processes large datasets in parallel batches

This job demonstrates:
- Kubernetes CronJob and Job patterns
- Parallel batch processing with work queues
- Database connection pooling for batch operations
- S3/object storage integration
- Idempotent processing with checkpointing
- Prometheus pushgateway for metrics
- Graceful shutdown handling
"""

from contextlib import suppress
import asyncio
import asyncpg
import boto3
import logging
import signal
import sys
import time
import os
import re
from typing import List, Dict, Any, Optional
from datetime import datetime, timezone
from prometheus_client import Counter, Histogram, Gauge, push_to_gateway
from prometheus_client import CollectorRegistry

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Configuration from environment
DATABASE_URL = os.environ.get("DATABASE_URL")
if not DATABASE_URL:
    raise RuntimeError("DATABASE_URL environment variable is required")
if not DATABASE_URL.startswith(("postgresql://", "postgres://")):
    raise RuntimeError("DATABASE_URL must use postgresql:// or postgres:// scheme")

S3_BUCKET = os.getenv("S3_BUCKET", "mantl-data")
if not S3_BUCKET:
    raise RuntimeError("S3_BUCKET environment variable must not be empty")
if not re.fullmatch(r"[a-z0-9][a-z0-9.\-]{1,61}[a-z0-9]", S3_BUCKET):
    raise RuntimeError(f"S3_BUCKET name is invalid: {S3_BUCKET!r}")


def _parse_int_env(name: str, default: int, min_val: int, max_val: int) -> int:
    raw = os.getenv(name, str(default))
    try:
        value = int(raw)
    except ValueError:
        raise RuntimeError(f"{name} must be a valid integer, got: {raw!r}") from None
    return max(min_val, min(value, max_val))


BATCH_SIZE = _parse_int_env("BATCH_SIZE", 1000, 1, 50000)
PARALLEL_WORKERS = _parse_int_env("PARALLEL_WORKERS", 4, 1, 32)
PUSHGATEWAY_URL = os.getenv("PUSHGATEWAY_URL", "http://prometheus-pushgateway:9091")
JOB_NAME = os.getenv("JOB_NAME", "batch-processor")
if not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_\-]*", JOB_NAME):
    raise RuntimeError(f"JOB_NAME contains invalid characters: {JOB_NAME!r}")

# Prometheus metrics
registry = CollectorRegistry()

RECORDS_PROCESSED = Counter(
    "batch_records_processed_total", "Total records processed", ["status"], registry=registry
)

PROCESSING_DURATION = Histogram(
    "batch_processing_duration_seconds",
    "Time spent processing batches",
    ["batch_id"],
    registry=registry,
)

BATCH_SIZE_METRIC = Gauge(
    "batch_current_size", "Current batch size being processed", registry=registry
)

JOB_START_TIME = Gauge("batch_job_start_timestamp", "Job start timestamp", registry=registry)

JOB_COMPLETION_TIME = Gauge(
    "batch_job_completion_timestamp", "Job completion timestamp", registry=registry
)

# Global state for graceful shutdown
shutdown_requested = asyncio.Event()


async def wait_for_shutdown(timeout: float):
    """Wake immediately on shutdown, or resume after the backoff interval."""
    with suppress(TimeoutError):
        await asyncio.wait_for(shutdown_requested.wait(), timeout=timeout)


def signal_handler(signum, frame):
    """Handle shutdown signals gracefully"""
    logger.info(f"Received signal {signum}, initiating graceful shutdown...")
    shutdown_requested.set()


# Register signal handlers
signal.signal(signal.SIGTERM, signal_handler)
signal.signal(signal.SIGINT, signal_handler)


class BatchProcessor:
    """Main batch processing orchestrator"""

    def __init__(self):
        self.db_pool: Optional[asyncpg.Pool] = None
        self.s3_client = boto3.client("s3")
        self.checkpoint_key = (
            f"checkpoints/{JOB_NAME}/{datetime.now(timezone.utc).strftime('%Y-%m-%d')}.json"
        )

    async def initialize(self):
        """Initialize database connection pool"""
        logger.info("Initializing database connection pool...")
        self.db_pool = await asyncpg.create_pool(
            DATABASE_URL, min_size=2, max_size=PARALLEL_WORKERS + 2, command_timeout=60
        )
        logger.info("Database pool initialized")

    async def cleanup(self):
        """Cleanup resources"""
        if self.db_pool:
            await self.db_pool.close()
            logger.info("Database pool closed")

    def load_checkpoint(self) -> Optional[int]:
        """Load processing checkpoint from S3"""
        try:
            response = self.s3_client.get_object(Bucket=S3_BUCKET, Key=self.checkpoint_key)
            checkpoint_data = response["Body"].read().decode("utf-8")
            import json

            checkpoint = json.loads(checkpoint_data)
            logger.info(f"Loaded checkpoint: last_processed_id={checkpoint.get('last_id')}")
            return checkpoint.get("last_id")
        except self.s3_client.exceptions.NoSuchKey:
            logger.info("No checkpoint found, starting from beginning")
            return None
        except Exception as e:
            logger.warning(f"Failed to load checkpoint: {e}")
            return None

    def save_checkpoint(self, last_processed_id: int):
        """Save processing checkpoint to S3"""
        try:
            import json

            checkpoint = {
                "last_id": last_processed_id,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "job_name": JOB_NAME,
            }
            self.s3_client.put_object(
                Bucket=S3_BUCKET,
                Key=self.checkpoint_key,
                Body=json.dumps(checkpoint).encode("utf-8"),
            )
            logger.info(f"Saved checkpoint: last_id={last_processed_id}")
        except Exception:
            logger.exception("Failed to save checkpoint")

    async def fetch_batch(self, offset: int, limit: int) -> List[Dict[str, Any]]:
        """Fetch a batch of records to process"""
        async with self.db_pool.acquire() as conn:
            query = """
                SELECT id, data, created_at, status
                FROM batch_items
                WHERE id > $1 AND status = 'pending'
                ORDER BY id
                LIMIT $2
            """
            rows = await conn.fetch(query, offset, limit)
            return [dict(row) for row in rows]

    async def process_record(self, record: Dict[str, Any]) -> bool:
        """Process a single record"""
        try:
            # Simulate processing work
            # In production, this would be actual data transformation/analysis
            await asyncio.sleep(0.1)  # Simulate processing time

            # Example: Parse data, transform, and update
            record_id = record["id"]
            data = record["data"]

            # Update record status
            async with self.db_pool.acquire() as conn:
                await conn.execute(
                    """
                    UPDATE batch_items
                    SET status = 'processed',
                        processed_at = NOW(),
                        result = $1
                    WHERE id = $2
                    """,
                    f"Processed: {data}",
                    record_id,
                )

            RECORDS_PROCESSED.labels(status="success").inc()
            return True

        except Exception as e:
            logger.exception("Failed to process record")
            RECORDS_PROCESSED.labels(status="error").inc()

            # Mark record as failed
            try:
                async with self.db_pool.acquire() as conn:
                    await conn.execute(
                        """
                        UPDATE batch_items
                        SET status = 'failed',
                            error_message = $1
                        WHERE id = $2
                        """,
                        str(e),
                        record["id"],
                    )
            except Exception:
                logger.exception("Failed to mark record as failed")

            return False

    async def process_batch_worker(self, batch: List[Dict[str, Any]], worker_id: int):
        """Worker coroutine to process records in parallel"""
        logger.info(f"Worker {worker_id} processing {len(batch)} records")

        for record in batch:
            if shutdown_requested.is_set():
                logger.info(f"Worker {worker_id} shutting down gracefully")
                break

            await self.process_record(record)

        logger.info(f"Worker {worker_id} completed")

    async def run(self):
        """Main batch processing logic"""

        logger.info(f"Starting batch job: {JOB_NAME}")
        JOB_START_TIME.set_to_current_time()

        await self.initialize()

        try:
            # Load checkpoint
            last_processed_id = self.load_checkpoint() or 0
            logger.info(f"Starting from ID: {last_processed_id}")

            total_processed = 0
            batch_number = 0

            while not shutdown_requested.is_set():
                batch_number += 1
                batch_start = time.time()

                # Fetch next batch
                batch = await self.fetch_batch(last_processed_id, BATCH_SIZE)

                if not batch:
                    logger.info("No more records to process")
                    break

                BATCH_SIZE_METRIC.set(len(batch))
                logger.info(f"Processing batch #{batch_number}: {len(batch)} records")

                # Split batch among workers
                chunk_size = max(1, len(batch) // PARALLEL_WORKERS)
                chunks = [batch[i : i + chunk_size] for i in range(0, len(batch), chunk_size)]

                # Process chunks in parallel
                tasks = [
                    self.process_batch_worker(chunk, worker_id)
                    for worker_id, chunk in enumerate(chunks)
                    if chunk
                ]
                await asyncio.gather(*tasks)

                # Interrupted workers may leave pending records; preserve the previous
                # checkpoint so the next run can select them again.
                if shutdown_requested.is_set():
                    break

                # Update progress
                last_processed_id = batch[-1]["id"]
                total_processed += len(batch)

                # Save checkpoint
                self.save_checkpoint(last_processed_id)

                # Record batch metrics
                batch_duration = time.time() - batch_start
                PROCESSING_DURATION.labels(batch_id=str(batch_number)).observe(batch_duration)

                logger.info(
                    f"Batch #{batch_number} complete: "
                    f"{len(batch)} records in {batch_duration:.2f}s "
                    f"({len(batch) / batch_duration:.2f} records/s)"
                )

                # Small delay between batches to avoid overwhelming database
                await wait_for_shutdown(0.5)

            logger.info(f"Job complete: {total_processed} total records processed")
            JOB_COMPLETION_TIME.set_to_current_time()

            # Push metrics to Pushgateway
            try:
                push_to_gateway(PUSHGATEWAY_URL, job=JOB_NAME, registry=registry)
                logger.info("Metrics pushed to Pushgateway")
            except Exception:
                logger.exception("Failed to push metrics")

        except Exception:
            logger.exception("Job failed")
            RECORDS_PROCESSED.labels(status="error").inc()
            raise

        finally:
            await self.cleanup()

        # Exit with appropriate code
        sys.exit(0 if not shutdown_requested.is_set() else 130)


async def setup_database_schema():
    """Create database schema if it doesn't exist (for demo purposes)"""
    conn = await asyncpg.connect(DATABASE_URL)
    try:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS batch_items (
                id SERIAL PRIMARY KEY,
                data TEXT NOT NULL,
                status VARCHAR(20) DEFAULT 'pending',
                created_at TIMESTAMP DEFAULT NOW(),
                processed_at TIMESTAMP,
                result TEXT,
                error_message TEXT
            );

            CREATE INDEX IF NOT EXISTS idx_batch_items_status
            ON batch_items(status, id);
        """)
        logger.info("Database schema ready")
    finally:
        await conn.close()


async def seed_test_data(count: int = 10000):
    """Seed database with test data (for demo purposes)"""
    conn = await asyncpg.connect(DATABASE_URL)
    try:
        # Check if data already exists
        existing = await conn.fetchval("SELECT COUNT(*) FROM batch_items WHERE status = 'pending'")
        if existing > 0:
            logger.info(f"Database already has {existing} pending records")
            return

        logger.info(f"Seeding {count} test records...")
        await conn.executemany(
            "INSERT INTO batch_items (data) VALUES ($1)",
            [(f"Test data item {i}",) for i in range(count)],
        )
        logger.info(f"Seeded {count} records")
    finally:
        await conn.close()


async def main():
    """Main entry point"""
    mode = os.getenv("MODE", "process")

    if mode == "setup":
        # Setup mode: Create schema and seed data
        await setup_database_schema()
        await seed_test_data()
    elif mode == "process":
        # Process mode: Run batch job
        processor = BatchProcessor()
        await processor.run()
    else:
        logger.error(f"Unknown mode: {mode}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
