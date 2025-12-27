.PHONY: validate kyverno-validate

validate: kyverno-validate ## Validate policies

kyverno-validate: ## Validate rendered manifests against Kyverno policies
	@kustomize build ../clusters/production > /tmp/rendered.yaml
	@docker run --rm -v "$(PWD)/..":/project -w /project ghcr.io/kyverno/kyverno:v1.12.1 kyverno apply policies/kyverno --resource /tmp/rendered.yaml --policy-report --skip-background-request
