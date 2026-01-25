# Tiltfile for Mantl Platform Development
# This file enables hot-reload development for all example applications

# Configuration
config.define_string("cluster", args=False, usage="Kubernetes cluster context to use (default: kind-mantl-dev)")
config.define_bool("build-images", args=True, usage="Build container images locally (default: True)")
cfg = config.parse()

# Set default cluster
allow_k8s_contexts(['kind-mantl-dev', 'docker-desktop', 'minikube'])

# ==============================================================================
# E-Commerce Microservices
# ==============================================================================

# Products API (Python/FastAPI)
docker_build(
    'mantl/products-api',
    context='./examples/ecommerce-microservices/src/products-api',
    dockerfile='./examples/ecommerce-microservices/src/products-api/Dockerfile',
    live_update=[
        sync('./examples/ecommerce-microservices/src/products-api', '/app'),
        run('pip install -r requirements.txt', trigger='./examples/ecommerce-microservices/src/products-api/requirements.txt'),
    ]
)

k8s_yaml([
    'examples/ecommerce-microservices/database/postgres-cluster.yaml',
    'examples/ecommerce-microservices/base/products-api/deployment.yaml',
    'examples/ecommerce-microservices/base/products-api/service.yaml',
])

k8s_resource(
    'products-api',
    port_forwards='8000:8000',
    labels=['ecommerce'],
    resource_deps=['postgres']
)

# Frontend (React/TypeScript)
docker_build(
    'mantl/ecommerce-frontend',
    context='./examples/ecommerce-microservices/src/frontend',
    dockerfile='./examples/ecommerce-microservices/src/frontend/Dockerfile.dev',
    live_update=[
        sync('./examples/ecommerce-microservices/src/frontend/src', '/app/src'),
        sync('./examples/ecommerce-microservices/src/frontend/public', '/app/public'),
    ]
)

k8s_yaml([
    'examples/ecommerce-microservices/base/frontend/deployment.yaml',
    'examples/ecommerce-microservices/base/frontend/service.yaml',
])

k8s_resource(
    'ecommerce-frontend',
    port_forwards='3000:3000',
    labels=['ecommerce']
)

# ==============================================================================
# ML Inference Service
# ==============================================================================

docker_build(
    'mantl/ml-inference',
    context='./examples/ml-inference-service/src',
    dockerfile='./examples/ml-inference-service/src/Dockerfile',
    live_update=[
        sync('./examples/ml-inference-service/src', '/app'),
        run('pip install -r requirements.txt', trigger='./examples/ml-inference-service/src/requirements.txt'),
    ]
)

k8s_yaml('examples/ml-inference-service/k8s/deployment.yaml')

k8s_resource(
    'ml-inference',
    port_forwards='8001:8000',
    labels=['ml']
)

# ==============================================================================
# Event-Driven Architecture
# ==============================================================================

# NATS JetStream
k8s_yaml('examples/event-driven-arch/k8s/nats.yaml')

k8s_resource(
    'nats',
    port_forwards=['4222:4222', '8222:8222'],
    labels=['messaging']
)

# Publisher
docker_build(
    'mantl/event-publisher',
    context='./examples/event-driven-arch/publisher',
    dockerfile='./examples/event-driven-arch/publisher/Dockerfile',
    live_update=[
        sync('./examples/event-driven-arch/publisher', '/app'),
        run('pip install -r requirements.txt', trigger='./examples/event-driven-arch/publisher/requirements.txt'),
    ]
)

k8s_yaml('examples/event-driven-arch/k8s/publisher.yaml')

k8s_resource(
    'event-publisher',
    port_forwards='8002:8000',
    labels=['messaging'],
    resource_deps=['nats']
)

# Subscriber
docker_build(
    'mantl/event-subscriber',
    context='./examples/event-driven-arch/subscriber',
    dockerfile='./examples/event-driven-arch/subscriber/Dockerfile',
    live_update=[
        sync('./examples/event-driven-arch/subscriber', '/app'),
        run('pip install -r requirements.txt', trigger='./examples/event-driven-arch/subscriber/requirements.txt'),
    ]
)

k8s_yaml('examples/event-driven-arch/k8s/subscriber.yaml')

k8s_resource(
    'event-subscriber',
    labels=['messaging'],
    resource_deps=['nats']
)

# ==============================================================================
# Static Website
# ==============================================================================

docker_build(
    'mantl/static-website',
    context='./examples/static-website',
    dockerfile='./examples/static-website/Dockerfile',
    live_update=[
        sync('./examples/static-website/src', '/usr/share/nginx/html'),
        sync('./examples/static-website/nginx.conf', '/etc/nginx/nginx.conf'),
    ]
)

k8s_yaml('examples/static-website/k8s/deployment.yaml')

k8s_resource(
    'mantl-website',
    port_forwards='8080:8080',
    labels=['frontend']
)

# ==============================================================================
# Platform Services (Optional - for full platform development)
# ==============================================================================

# Uncomment to deploy full platform stack
# k8s_yaml(helm(
#     './platform/helm-charts/mantl-platform',
#     name='mantl-platform',
#     namespace='mantl-system',
#     values=['./platform/helm-charts/mantl-platform/values-dev.yaml']
# ))

# ==============================================================================
# Helper Functions
# ==============================================================================

# Watch for changes in Kubernetes manifests
watch_file('examples/ecommerce-microservices/base/')
watch_file('examples/ml-inference-service/k8s/')
watch_file('examples/event-driven-arch/k8s/')
watch_file('examples/static-website/k8s/')

# Print helpful development URLs
print("""
╔══════════════════════════════════════════════════════════════╗
║              Mantl Platform Development Mode                 ║
╚══════════════════════════════════════════════════════════════╝

🚀 Services running:

  E-Commerce:
    📦 Products API:       http://localhost:8000
    🎨 Frontend:           http://localhost:3000

  ML:
    🤖 ML Inference:       http://localhost:8001
    📊 Prediction API:     http://localhost:8001/docs

  Messaging:
    📨 Event Publisher:    http://localhost:8002
    🔄 NATS Monitor:       http://localhost:8222
    📡 NATS Client:        nats://localhost:4222

  Static:
    🌐 Website:            http://localhost:8080

💡 Tips:
  - Press 's' to open Tilt UI in browser
  - Press 'space' to select a resource
  - Press 'r' to trigger a rebuild
  - Press 'q' to quit Tilt

📝 Logs are automatically tailed in the Tilt UI

""")
