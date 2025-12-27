Spinnaker add-on (Operator + SpinnakerService)

Overview
This add-on installs/configures Spinnaker via the Spinnaker Operator and a SpinnakerService CR. Use overlays to choose storage (S3/GCS) and to wire Kubernetes accounts per environment. If managing via ArgoCD, use `clusters/production/apps/spinnaker.yaml`. Ensure the operator is installed first (see operator/README.md).

Prerequisites
- A Kubernetes cluster reachable by kubectl
- Spinnaker Operator installed in the cluster (see operator docs)
- An object storage bucket (e.g., S3 or GCS) and credentials

Layout
- base/: Namespaces and a minimal SpinnakerService template with placeholders
- overlays/s3: Patch to enable S3 storage (Front50)
- overlays/gcs: Patch to enable GCS storage (Front50)
- accounts/: Example RBAC + service accounts Spinnaker will use to manage target namespaces

Quick start
1) Install the Operator (follow upstream docs), then create namespaces:
   kubectl apply -k addons/spinnaker/base

2) Choose a storage overlay:
   # S3
   kubectl apply -k addons/spinnaker/overlays/s3
   # or GCS
   kubectl apply -k addons/spinnaker/overlays/gcs

3) Add Kubernetes accounts (service accounts + roles) in your target namespaces, and reference them in the SpinnakerService spec (providers.kubernetes.accounts).

4) Expose Deck (UI) and Gate (API) using your ingress controller or change the Service type in spec.spinnakerConfig.config.services.

CI triggers
- Jenkins: use the native Jenkins trigger and publish build.properties (see examples/ci/jenkins/Jenkinsfile).
- Generic webhook: POST to Gate with artifact refs (see examples/ci/webhook/spinnaker-trigger.sh).

Canary (Kayenta)
- Enable later once Prometheus/Datadog/Google Cloud Monitoring is available and SLOs are defined.

Notes
- All values with TODO_ must be set before applying.
- If you prefer Halyard, keep this directory but generate halconfig and deploy separately.
