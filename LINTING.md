# Linting Issues in Mantl

This document tracks the current linting issues in the Mantl repository that need to be addressed.

## Python Linting Issues

The repository has several Python files with linting issues that need to be fixed:

- Style issues (PEP8): whitespace, line length, blank lines
- Import order issues
- Unused imports
- Function and class definition spacing
- Bare except clauses
- Invalid escape sequences in regular expressions
- Syntax errors in older Python code that needs updating

## Fixing Linting Issues

To fix the linting issues, run:

```bash
# Install required tools
pip install -r requirements.txt

# Run the linting script to see all issues
sh scripts/lint.sh

# Fix issues manually or use autopep8 for simple fixes
pip install autopep8
autopep8 --in-place --aggressive --aggressive <path-to-file>
```

## Priority Files to Fix

The highest priority files to fix are:

1. Library files in `library/` directory
2. Plugin files in `plugins/` directory
3. Test files in `tests/` directory
4. Documentation configuration in `docs/conf.py`

## Tracking Progress

As linting issues are fixed, please update this document to track progress.

## Linting in CI/CD

The Jenkinsfile includes a linting stage that runs flake8, pylint, and ansible-lint. Currently, this stage may fail due to existing issues. Once linting issues are fixed, this stage will be a required check for all PRs.