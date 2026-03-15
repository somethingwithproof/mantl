"""
Products API - E-Commerce Microservices Example
Fast API service for product catalog management
"""
from fastapi import FastAPI, HTTPException, Query, Depends
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import Counter, Histogram, make_asgi_app
from opentelemetry import trace
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from pydantic import BaseModel, Field
from typing import List, Optional
import asyncpg
import os
import logging
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Prometheus metrics
REQUEST_COUNT = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)
REQUEST_DURATION = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration',
    ['method', 'endpoint']
)
PRODUCT_OPERATIONS = Counter(
    'product_operations_total',
    'Total product operations',
    ['operation', 'status']
)

# OpenTelemetry tracer
tracer = trace.get_tracer(__name__)

# Database pool
db_pool = None

# Pydantic models
class Product(BaseModel):
    id: Optional[int] = None
    name: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=1000)
    price: float = Field(..., gt=0)
    category: str = Field(..., min_length=1, max_length=100)
    stock: int = Field(..., ge=0)
    sku: str = Field(..., min_length=1, max_length=50)
    image_url: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None

    class Config:
        json_schema_extra = {
            "example": {
                "name": "Wireless Mouse",
                "description": "Ergonomic wireless mouse with 2.4GHz connection",
                "price": 29.99,
                "category": "Electronics",
                "stock": 150,
                "sku": "MOUSE-001",
                "image_url": "https://example.com/images/mouse.jpg"
            }
        }

class ProductCreate(BaseModel):
    name: str
    description: Optional[str] = None
    price: float
    category: str
    stock: int
    sku: str
    image_url: Optional[str] = None

class ProductUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    price: Optional[float] = None
    category: Optional[str] = None
    stock: Optional[int] = None
    sku: Optional[str] = None
    image_url: Optional[str] = None

# FastAPI app
app = FastAPI(
    title="Products API",
    description="E-Commerce product catalog management",
    version="1.0.0",
    docs_url="/api/v1/docs",
    redoc_url="/api/v1/redoc",
    openapi_url="/api/v1/openapi.json"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure properly in production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Prometheus metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# OpenTelemetry instrumentation
FastAPIInstrumentor.instrument_app(app)

# Database dependency
async def get_db():
    """Get database connection from pool"""
    global db_pool
    if db_pool is None:
        raise HTTPException(status_code=503, detail="Database not initialized")
    async with db_pool.acquire() as connection:
        yield connection

# Startup/Shutdown events
@app.on_event("startup")
async def startup():
    """Initialize database connection pool"""
    global db_pool
    database_url = os.getenv("DATABASE_URL")
    if not database_url:
        raise RuntimeError("DATABASE_URL environment variable not set")

    logger.info("Initializing database connection pool")
    db_pool = await asyncpg.create_pool(database_url, min_size=5, max_size=20)

    # Create table if not exists
    async with db_pool.acquire() as conn:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS products (
                id SERIAL PRIMARY KEY,
                name VARCHAR(200) NOT NULL,
                description TEXT,
                price DECIMAL(10, 2) NOT NULL,
                category VARCHAR(100) NOT NULL,
                stock INTEGER NOT NULL DEFAULT 0,
                sku VARCHAR(50) UNIQUE NOT NULL,
                image_url TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );

            CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
            CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
        """)

    logger.info("Database initialized successfully")

@app.on_event("shutdown")
async def shutdown():
    """Close database connection pool"""
    global db_pool
    if db_pool:
        logger.info("Closing database connection pool")
        await db_pool.close()

# Health check endpoints
@app.get("/health")
async def health():
    """Liveness probe"""
    return {"status": "healthy"}

@app.get("/ready")
async def ready():
    """Readiness probe"""
    global db_pool
    if db_pool is None:
        raise HTTPException(status_code=503, detail="Database not ready")

    try:
        async with db_pool.acquire() as conn:
            await conn.fetchval("SELECT 1")
        return {"status": "ready"}
    except Exception as e:
        logger.error(f"Readiness check failed: {e}")
        raise HTTPException(status_code=503, detail="Database connection failed")

# API endpoints
@app.get("/api/v1/products", response_model=List[Product])
async def list_products(
    category: Optional[str] = Query(None, description="Filter by category"),
    min_price: Optional[float] = Query(None, ge=0, description="Minimum price"),
    max_price: Optional[float] = Query(None, ge=0, description="Maximum price"),
    limit: int = Query(100, ge=1, le=1000, description="Number of products to return"),
    offset: int = Query(0, ge=0, description="Number of products to skip"),
    db=Depends(get_db)
):
    """List all products with optional filtering"""
    with tracer.start_as_current_span("list_products"):
        query = "SELECT * FROM products WHERE 1=1"
        params = []
        param_index = 1

        if category:
            query += f" AND category = ${param_index}"
            params.append(category)
            param_index += 1

        if min_price is not None:
            query += f" AND price >= ${param_index}"
            params.append(min_price)
            param_index += 1

        if max_price is not None:
            query += f" AND price <= ${param_index}"
            params.append(max_price)
            param_index += 1

        query += f" ORDER BY created_at DESC LIMIT ${param_index} OFFSET ${param_index + 1}"
        params.extend([limit, offset])

        try:
            rows = await db.fetch(query, *params)
            PRODUCT_OPERATIONS.labels(operation='list', status='success').inc()
            return [dict(row) for row in rows]
        except Exception as e:
            logger.error(f"Failed to list products: {e}")
            PRODUCT_OPERATIONS.labels(operation='list', status='error').inc()
            raise HTTPException(status_code=500, detail="Failed to list products")

@app.get("/api/v1/products/{product_id}", response_model=Product)
async def get_product(product_id: int, db=Depends(get_db)):
    """Get a specific product by ID"""
    with tracer.start_as_current_span("get_product"):
        try:
            row = await db.fetchrow(
                "SELECT * FROM products WHERE id = $1",
                product_id
            )
            if row is None:
                PRODUCT_OPERATIONS.labels(operation='get', status='not_found').inc()
                raise HTTPException(status_code=404, detail="Product not found")

            PRODUCT_OPERATIONS.labels(operation='get', status='success').inc()
            return dict(row)
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get product {product_id}: {e}")
            PRODUCT_OPERATIONS.labels(operation='get', status='error').inc()
            raise HTTPException(status_code=500, detail="Failed to get product")

@app.post("/api/v1/products", response_model=Product, status_code=201)
async def create_product(product: ProductCreate, db=Depends(get_db)):
    """Create a new product"""
    with tracer.start_as_current_span("create_product"):
        try:
            row = await db.fetchrow("""
                INSERT INTO products (name, description, price, category, stock, sku, image_url)
                VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING *
            """, product.name, product.description, product.price, product.category,
                product.stock, product.sku, product.image_url)

            PRODUCT_OPERATIONS.labels(operation='create', status='success').inc()
            logger.info(f"Created product: {product.sku}")
            return dict(row)
        except asyncpg.UniqueViolationError:
            PRODUCT_OPERATIONS.labels(operation='create', status='duplicate').inc()
            raise HTTPException(status_code=409, detail="Product with this SKU already exists")
        except Exception as e:
            logger.error(f"Failed to create product: {e}")
            PRODUCT_OPERATIONS.labels(operation='create', status='error').inc()
            raise HTTPException(status_code=500, detail="Failed to create product")

@app.put("/api/v1/products/{product_id}", response_model=Product)
async def update_product(product_id: int, product: ProductUpdate, db=Depends(get_db)):
    """Update an existing product"""
    with tracer.start_as_current_span("update_product"):
        # Build dynamic update query
        updates = []
        params = []
        param_index = 1

        for field, value in product.dict(exclude_unset=True).items():
            if value is not None:
                updates.append(f"{field} = ${param_index}")
                params.append(value)
                param_index += 1

        if not updates:
            raise HTTPException(status_code=400, detail="No fields to update")

        params.append(product_id)
        query = f"""
            UPDATE products
            SET {', '.join(updates)}, updated_at = CURRENT_TIMESTAMP
            WHERE id = ${param_index}
            RETURNING *
        """

        try:
            row = await db.fetchrow(query, *params)
            if row is None:
                PRODUCT_OPERATIONS.labels(operation='update', status='not_found').inc()
                raise HTTPException(status_code=404, detail="Product not found")

            PRODUCT_OPERATIONS.labels(operation='update', status='success').inc()
            logger.info(f"Updated product: {product_id}")
            return dict(row)
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to update product {product_id}: {e}")
            PRODUCT_OPERATIONS.labels(operation='update', status='error').inc()
            raise HTTPException(status_code=500, detail="Failed to update product")

@app.delete("/api/v1/products/{product_id}", status_code=204)
async def delete_product(product_id: int, db=Depends(get_db)):
    """Delete a product"""
    with tracer.start_as_current_span("delete_product"):
        try:
            result = await db.execute(
                "DELETE FROM products WHERE id = $1",
                product_id
            )
            if result == "DELETE 0":
                PRODUCT_OPERATIONS.labels(operation='delete', status='not_found').inc()
                raise HTTPException(status_code=404, detail="Product not found")

            PRODUCT_OPERATIONS.labels(operation='delete', status='success').inc()
            logger.info(f"Deleted product: {product_id}")
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to delete product {product_id}: {e}")
            PRODUCT_OPERATIONS.labels(operation='delete', status='error').inc()
            raise HTTPException(status_code=500, detail="Failed to delete product")

@app.get("/api/v1/categories", response_model=List[str])
async def list_categories(db=Depends(get_db)):
    """List all product categories"""
    with tracer.start_as_current_span("list_categories"):
        try:
            rows = await db.fetch(
                "SELECT DISTINCT category FROM products ORDER BY category"
            )
            return [row['category'] for row in rows]
        except Exception as e:
            logger.error(f"Failed to list categories: {e}")
            raise HTTPException(status_code=500, detail="Failed to list categories")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        log_level=os.getenv("LOG_LEVEL", "info").lower()
    )
