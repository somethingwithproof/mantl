#!/usr/bin/env bash
set -euo pipefail

NAME=${1:?usage: $0 <service-name>}
DST="applications/services/${NAME}"

if [ -e "$DST" ]; then
  echo "Destination $DST already exists" >&2
  exit 1
fi

mkdir -p "$(dirname "$DST")"
cp -R applications/templates/web-service/base "$DST"
sed -i.bak "s/name: app/name: ${NAME}/g" "$DST"/deployment.yaml && rm "$DST"/deployment.yaml.bak
sed -i.bak "s/name: app/name: ${NAME}/g" "$DST"/service.yaml && rm "$DST"/service.yaml.bak

cat <<EOF
Created $DST
Next steps:
  - Update image and domain in $DST/*
  - Deploy with: kubectl apply -k $DST
EOF
