#!/bin/bash
set -e

echo "Setting up Kubernetes-Nomad integration tests..."

# Setup SSH keys if they don't exist
if [ ! -f config/.ssh/id_rsa ]; then
  echo "Generating SSH keys..."
  bash scripts/setup-ssh.sh
fi

echo "Building and starting test environment..."
docker-compose build
docker-compose up -d

echo "Waiting for environment to initialize (this may take a few minutes)..."
sleep 60

echo "Checking container status..."
docker-compose ps

echo "Running tests..."
docker-compose exec test pytest -v

echo "Tests complete!"
echo "To view detailed results, check the results directory."
echo "To clean up the test environment: docker-compose down -v"