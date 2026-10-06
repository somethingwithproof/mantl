# SPDX-License-Identifier: Apache-2.0
# Multi-stage build for ML inference service
# Official Python 3.12.12 runtime; NumPy 2.5 requires Python >= 3.12.
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c AS builder

WORKDIR /build

# Copy and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --only-binary :all: --require-hashes --user -r requirements.txt

# Final stage
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c

# Create non-root user
RUN groupadd -r mluser && useradd -r -g mluser -u 1000 mluser

WORKDIR /app

# Copy Python packages from builder
COPY --from=builder /root/.local /home/mluser/.local
ENV PYTHONDONTWRITEBYTECODE=1
ENV PATH=/home/mluser/.local/bin:$PATH

# Copy application code
COPY main.py .

# Create model directory
RUN mkdir -p /app/models && chown mluser:mluser /app/models

USER mluser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=3)"

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
