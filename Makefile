# Mantl developer Makefile (legacy baseline)
# This file will be modernized (ruff/black, format targets, improved linting).
# For now, we add human-readable comments without changing behavior.

.PHONY: all docs test lint clean

# Default target: run tests, lint, and docs.
all: test lint docs

# Install Python dependencies for local dev.
deps:
	pip install -r requirements.txt

# Build the Sphinx documentation (outputs to docs/_build/html).
docs:
	cd docs && make html

# Run Python tests (pytest discovers tests in tests/ per setup.cfg).
test:
	pytest tests/

# Lint: Python via ruff/black (check only) and Ansible roles via ansible-lint.
# Note: adjust paths if playbooks/ is introduced.
lint:
	black --check .
	ruff .
	ansible-lint roles/*

# Remove common build artifacts and caches.
clean:
	rm -rf docs/_build
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

# Format all Terraform files in-place. Consider using -recursive.
terraform-fmt:
	find . -name "*.tf" -exec terraform fmt {} \;

# Validate Terraform modules in the terraform/ tree without hitting backends.
terraform-validate:
	find terraform -name "*.tf" -not -path "*/\.*" -exec sh -c "cd \$$(dirname {}) && terraform validate" \;

# Show available targets and descriptions.
# Autoformat source code (Python and Terraform)
format:
	black .
	ruff --fix .
	find . -name "*.tf" -exec terraform fmt {} \;

yamllint:
	yamllint .

# Release helpers (requires GitHub CLI and proper permissions)
tag:
	@if [ -z "$(VERSION)" ]; then echo "Set VERSION, e.g. make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

release:
	@if [ -z "$(VERSION)" ]; then echo "Set VERSION, e.g. make release VERSION=v1.0.0"; exit 1; fi
	gh release create $(VERSION) --generate-notes || true

protect:
	./scripts/set-required-checks.sh main

# Bootstrap helper
bootstrap:
	./scripts/bootstrap-cluster.sh production aws

help:
	@echo "Available targets:"
	@echo "  all            - Run tests, linting, and build docs"
	@echo "  deps           - Install dependencies"
	@echo "  docs           - Build documentation"
	@echo "  test           - Run tests"
	@echo "  lint           - Check code quality"
	@echo "  format         - Autoformat Python and Terraform files"
	@echo "  yamllint       - Lint YAML files"
	@echo "  clean          - Remove build artifacts"
	@echo "  terraform-fmt  - Format Terraform files"
	@echo "  terraform-validate - Validate Terraform syntax"
	@echo "  tag            - Create and push a git tag (VERSION=vX.Y.Z)"
	@echo "  release        - Create a GitHub Release (VERSION=vX.Y.Z)"
	@echo "  protect        - Configure required checks on main via GitHub CLI"
