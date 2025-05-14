#!/bin/bash
# Lint script for Mantl
# Runs various linting tools on the codebase

set -e

echo "Running Python linting..."
flake8 . --exclude=.git,docs/_build,docs/source/conf.py,old,build,dist

echo "Running Ansible linting..."
find ./roles -name "*.yml" -print0 | xargs -0 ansible-lint
find ./playbooks -name "*.yml" -print0 | xargs -0 ansible-lint

echo "Running Terraform formatting check..."
terraform_files=$(find . -name "*.tf" -not -path "*/\.*")
for file in $terraform_files; do
  terraform fmt -check=true "$file" || {
    echo "Error: Terraform file $file is not properly formatted."
    echo "Run 'terraform fmt' to fix."
    exit 1
  }
done

echo "All linting passed!"