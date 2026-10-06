#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
mise trust --yes
mise install go python terraform kubectl kustomize helm uv
mise exec -- python -m venv .venv
mise exec -- .venv/bin/python -m pip install --only-binary :all: --require-hashes \
  -r requirements.txt -r requirements-test.txt
mise exec -- .venv/bin/pre-commit install

# Cluster creation remains an explicit developer choice.
if [[ "${CREATE_KIND_CLUSTER:-false}" == "true" ]]; then
  if ! command -v kind >/dev/null 2>&1; then
    echo "Install a pinned kind version with mise before requesting a local cluster." >&2
    exit 1
  fi
  kind create cluster --name mantl-dev --config=- <<'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
EOF
fi

printf '%s\n' "Mantl development tools and Python dependencies are ready."
