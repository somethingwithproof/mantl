package k8s.labels

deny[msg] {
  input.kind == "Deployment"
  not input.metadata.labels["app.kubernetes.io/name"]
  msg := sprintf("%s: missing app.kubernetes.io/name label", [input.metadata.name])
}
