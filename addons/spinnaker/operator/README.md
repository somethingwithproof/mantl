Spinnaker Operator Install
==========================

Use the upstream Spinnaker Operator (CRDs + controller). Recommended approach:

1. Pick a specific operator version.
2. Apply CRDs and the operator controller to the management cluster/namespace `spinnaker-operator`.
3. Manage SpinnakerService resources in namespace `spinnaker`.

Example (replace VERSION with a pinned tag):

kubectl apply -f https://raw.githubusercontent.com/armory/spinnaker-operator/VERSION/deploy/crds/all.yaml
kubectl apply -n spinnaker-operator -f https://raw.githubusercontent.com/armory/spinnaker-operator/VERSION/deploy/operator/armory-operator.yaml

Then apply the SpinnakerService from base/ or cloud overlays.