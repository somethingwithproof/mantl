#!/bin/bash
# Test script for Mantl
# Runs tests and generates coverage reports

set -e

# Run unit tests with coverage
echo "Running unit tests..."
pytest tests/ --cov=. --cov-report=term --cov-report=html:coverage_html

# Output test summary
echo "Test coverage report:"
echo "For detailed report, see coverage_html/index.html"