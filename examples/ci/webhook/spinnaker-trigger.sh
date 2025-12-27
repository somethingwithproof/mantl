#!/usr/bin/env bash
set -euo pipefail

GATE_URL=${GATE_URL:-https://spinnaker.example.com}
WEBHOOK=${WEBHOOK:-app-deploy}
IMAGE=${IMAGE:?set IMAGE to the pushed image tag}
ENVIRONMENT=${ENVIRONMENT:-staging}

curl -sSf -X POST -H "Content-Type: application/json" \
  -d "{\"artifacts\":[{\"type\":\"docker/image\",\"name\":\"${IMAGE}\"}], \"parameters\":{\"deployEnv\":\"${ENVIRONMENT}\"}}" \
  "${GATE_URL}/webhooks/webhook/${WEBHOOK}"
