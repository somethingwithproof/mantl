# Backstage Integration

## Import this repo
- In Backstage, use "Register Existing Component" and paste the URL to `catalog-info.yaml` in this repo.

## Use the scaffolder template
- Add `templates/backstage/web-service/template.yaml` to your Backstage Scaffolder catalog.
- Generate a new service, then commit the generated `applications/services/<name>` directory.

## CLI alternative
- Use `scripts/create-service.sh <service-name>` to copy the web-service template locally.
