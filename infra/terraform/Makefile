# SPDX-License-Identifier: Apache-2.0
.PHONY: validate fmt tflint tfsec checkov

validate: ## terraform init/validate for all blueprints
	@for dir in blueprints/*/; do \
		echo "Validating $$dir"; \
		cd "$$dir" && terraform init -backend=false -input=false && terraform validate && cd - >/dev/null; \
	done

fmt: ## terraform fmt recursively
	@terraform fmt -recursive

tflint: ## run tflint (requires tflint --init once)
	@tflint --init || true
	@tflint -f compact

tfsec: ## run tfsec on infra/terraform
	@tfsec --config-file ../../.tfsec.yml . || true

checkov: ## run checkov on infra/terraform
	@checkov -d . -o cli || true
