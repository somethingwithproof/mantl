GitHub artifacts

Create a file named github.env in this directory with:
  token=ghp_your_personal_access_token

Scopes:
- repo (read access) is sufficient for public/private repos used by Spinnaker.

This token is loaded into a Kubernetes Secret (github-artifacts) via kustomize and referenced by Spinnaker using encrypted:k8s syntax.
