# Multi-stage build for ML inference service
# Official Python 3.12.12 runtime; NumPy 2.5 requires Python >= 3.12.
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc g++ && \
    rm -rf /var/lib/apt/lists/*

# Copy and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Final stage
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r mluser && useradd -r -g mluser -u 1000 mluser

WORKDIR /app

# Copy Python packages from builder
COPY --from=builder /root/.local /home/mluser/.local
ENV PATH=/home/mluser/.local/bin:$PATH

# Copy application code
COPY main.py .

# Create model directory
RUN mkdir -p /app/models && chown -R mluser:mluser /app

USER mluser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
