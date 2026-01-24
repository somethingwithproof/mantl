FROM python:3.11-slim

# Install dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r batchuser && useradd -r -g batchuser -u 1000 batchuser

WORKDIR /app

# Copy requirements and install
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY batch_processor.py .

# Set ownership
RUN chown -R batchuser:batchuser /app

USER batchuser

CMD ["python", "batch_processor.py"]
