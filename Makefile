.PHONY: all docs test lint clean

# Default target
all: test lint docs

# Install dependencies
deps:
	pip install -r requirements.txt

# Build documentation
docs:
	cd docs && make html

# Run tests
test:
	pytest tests/

# Check code quality
lint:
	flake8 .
	ansible-lint roles/* playbooks/*

# Clean build artifacts
clean:
	rm -rf docs/_build
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

# Format Terraform files
terraform-fmt:
	find . -name "*.tf" -exec terraform fmt {} \;

# Check Terraform syntax
terraform-validate:
	find terraform -name "*.tf" -not -path "*/\.*" -exec sh -c "cd \$$(dirname {}) && terraform validate" \;

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Run tests, linting, and build docs"
	@echo "  deps           - Install dependencies"
	@echo "  docs           - Build documentation"
	@echo "  test           - Run tests"
	@echo "  lint           - Check code quality"
	@echo "  clean          - Remove build artifacts"
	@echo "  terraform-fmt  - Format Terraform files"
	@echo "  terraform-validate - Validate Terraform syntax"