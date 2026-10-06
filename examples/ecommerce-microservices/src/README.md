<!-- SPDX-License-Identifier: Apache-2.0 -->
# E-Commerce Microservices - Source Code

Complete source code for the example e-commerce application demonstrating mantl platform capabilities.

## Applications

### Products API
**Location**: `products-api/`

Python FastAPI service managing product catalog.

**Features**:
- RESTful API for CRUD operations
- PostgreSQL database with asyncpg
- Prometheus metrics exposition
- OpenTelemetry distributed tracing
- Health check endpoints
- Connection pooling

**Build & Run**:
```bash
cd products-api

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export DATABASE_URL="postgresql://user:password@localhost:5432/products"
export LOG_LEVEL="info"

# Run locally
python main.py

# Build Docker image
docker build -t products-api:1.0.0 .

# Run container
docker run -p 8000:8000 \
  -e DATABASE_URL="postgresql://user:password@db:5432/products" \
  products-api:1.0.0
```

**API Endpoints**:
- `GET /api/v1/products` - List all products
- `GET /api/v1/products/{id}` - Get product by ID
- `POST /api/v1/products` - Create new product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product
- `GET /api/v1/categories` - List categories
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

### Frontend
**Location**: `frontend/`

React TypeScript SPA for the storefront.

**Features**:
- Product browsing and filtering
- Category navigation
- Responsive design
- Nginx reverse proxy
- Health check endpoints
- Prometheus metrics (via nginx-exporter sidecar)

**Build & Run**:
```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm start

# Build for production
npm run build

# Build Docker image
docker build -t frontend:1.0.0 .

# Run container
docker run -p 8080:8080 frontend:1.0.0
```

**Environment Variables**:
- `REACT_APP_API_URL` - Base URL for backend APIs (default: `/api`)

The frontend uses Vite with Node 24, pinned in `frontend/mise.toml`. Run
`mise trust && mise install` from that directory, then `mise exec -- npm ci`,
`mise exec -- npm test`, and `mise exec -- npm run build`. `npm start` serves the
demo on port 3000; `/api` requests proxy to `http://localhost:8000`, configurable
with `API_PROXY_TARGET`. `REACT_APP_API_URL` remains a public build-time setting
and can also be supplied as a Docker build argument. Do not put credentials in it.
Production output remains in `build/` and the Nginx container serves port 8080
as its existing non-root `nginx` user. GitHub Actions checks the dependency audit,
TypeScript build, dev/API behavior, and production container.

## Development Workflow

### Local Development

```bash
# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=products \
  -p 5432:5432 \
  postgres:15

# Start Products API
cd products-api
export DATABASE_URL="postgresql://postgres:password@localhost:5432/products"
python main.py

# Start Frontend (in another terminal)
cd frontend
npm start
```

Access the application at http://localhost:3000

### Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: password
      POSTGRES_DB: products
    ports:
      - "5432:5432"

  products-api:
    build: ./products-api
    environment:
      DATABASE_URL: postgresql://postgres:password@postgres:5432/products
    ports:
      - "8000:8000"
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    ports:
      - "8080:8080"
    depends_on:
      - products-api
```

## Testing

### Products API Tests

```bash
cd products-api

# Install test dependencies
pip install pytest pytest-asyncio httpx

# Run tests
pytest tests/
```

### Frontend Tests

```bash
cd frontend

# Run unit tests
npm test

# Run E2E tests
npm run test:e2e
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build and Push

on:
  push:
    branches: [ main ]
    paths:
    - 'examples/ecommerce-microservices/src/**'

jobs:
  build-products-api:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: examples/ecommerce-microservices/src/products-api
        push: true
        tags: ghcr.io/${{ github.repository }}/products-api:${{ github.sha }}

  build-frontend:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: examples/ecommerce-microservices/src/frontend
        push: true
        tags: ghcr.io/${{ github.repository }}/frontend:${{ github.sha }}
```

## Security Scanning

### Trivy Scanning

```bash
# Scan Products API image
trivy image products-api:1.0.0

# Scan Frontend image
trivy image frontend:1.0.0
```

### Snyk Scanning

```bash
# Scan Python dependencies
cd products-api
snyk test

# Scan npm dependencies
cd frontend
snyk test
```

## Monitoring & Observability

### Prometheus Metrics

Both services expose Prometheus metrics:

**Products API** - `http://localhost:8000/metrics`:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration histogram
- `product_operations_total` - Product CRUD operation counters

**Frontend** - `http://localhost:9113/metrics` (via nginx-exporter sidecar):
- `nginx_http_requests_total` - Total nginx requests
- `nginx_http_request_duration_seconds` - Request duration
- `nginx_connections_active` - Active connections

### Distributed Tracing

Products API includes OpenTelemetry instrumentation. Configure the exporter:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://tempo.mantl-system:4317"
export OTEL_SERVICE_NAME="products-api"
```

View traces in Jaeger/Tempo UI.

### Logging

Structured JSON logging:

```json
{
  "timestamp": "2024-01-24T10:30:00Z",
  "level": "INFO",
  "message": "Created product: MOUSE-001",
  "service": "products-api",
  "trace_id": "abc123...",
  "span_id": "def456..."
}
```

## Database Migrations

Products API uses Alembic for database migrations:

```bash
cd products-api

# Create migration
alembic revision --autogenerate -m "Add products table"

# Apply migrations
alembic upgrade head

# Rollback migration
alembic downgrade -1
```

## Performance Optimization

### Products API

- **Connection Pooling**: asyncpg pool (min=5, max=20)
- **Query Optimization**: Indexed columns (category, sku)
- **Caching**: Redis integration (optional)
- **Async I/O**: All database operations are async

### Frontend

- **Code Splitting**: React lazy loading
- **Asset Optimization**: Gzip compression, cache headers
- **CDN Ready**: Static assets with immutable cache
- **Image Optimization**: Lazy loading, responsive images

## Troubleshooting

### Products API

**Database Connection Issues**:
```bash
# Test connection
python -c "import asyncpg; asyncpg.connect('postgresql://...')"

# Check logs
docker logs products-api
```

**High Memory Usage**:
- Reduce connection pool size
- Check for connection leaks
- Monitor with Prometheus metrics

### Frontend

**API Connection Issues**:
```bash
# Check nginx proxy configuration
docker exec frontend cat /etc/nginx/nginx.conf

# Test API connectivity
docker exec frontend wget -O- http://products-api:8000/health
```

**Build Failures**:
```bash
# Clear npm cache
npm cache clean --force

# Remove node_modules and rebuild
rm -rf node_modules package-lock.json
npm install
```

## Contributing

See main [CONTRIBUTING.md](../../../CONTRIBUTING.md)

## License

Apache 2.0
