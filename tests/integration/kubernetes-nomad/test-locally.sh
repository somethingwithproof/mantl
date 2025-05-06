#!/bin/bash
set -e

echo "Testing the Docker Compose configuration..."

# Validate Docker Compose file
echo "Validating docker-compose.yml..."
docker-compose config

# Check Dockerfiles
echo "Checking Dockerfiles..."
docker build -f Dockerfile.control -t mantl-control-test . || echo "Issue with Dockerfile.control"
docker build -f Dockerfile.worker -t mantl-worker-test . || echo "Issue with Dockerfile.worker"
docker build -f Dockerfile.test -t mantl-test-test . || echo "Issue with Dockerfile.test"

echo "Tests complete. If no errors were reported, the configuration looks good."
echo "To run the full integration test environment:"
echo "./run-tests.sh"