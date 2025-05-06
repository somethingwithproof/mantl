#!/bin/bash
set -e

echo "Starting integration tests for Kubernetes-Nomad integration..."

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 120

# Check if control node is up
echo "Checking connectivity to control node..."
ping -c 3 $CONTROL_NODE

# Set up environment for tests
export NOMAD_ADDR="http://${CONTROL_NODE}:4646"
export CONSUL_HTTP_ADDR="http://${CONTROL_NODE}:8500"
export VAULT_ADDR="http://${CONTROL_NODE}:8200"
export VAULT_TOKEN="root_token_for_testing"

# Add control node to known hosts to avoid SSH prompts
mkdir -p ~/.ssh
echo "${CONTROL_NODE} $(ssh-keyscan -t rsa ${CONTROL_NODE} 2>/dev/null)" >> ~/.ssh/known_hosts

# Wait for Kubernetes to be ready
echo "Waiting for Kubernetes API..."
until kubectl cluster-info; do
  sleep 5
done

# Wait for Nomad to be ready
echo "Waiting for Nomad API..."
until nomad server members; do
  sleep 5
done

echo "Waiting for example job to be deployed..."
sleep 60

# Run the tests
echo "Running integration tests..."
pytest -v --html=/etc/mantl/results/report.html

echo "Tests complete. See results in /etc/mantl/results/report.html"