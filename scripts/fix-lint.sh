#!/bin/bash
# Script to automatically fix common Python linting issues

set -e

echo "Installing required tools..."
pip install autopep8 isort

echo "Fixing common Python style issues with autopep8..."
# Fix Python files with autopep8
find . -name "*.py" -not -path "*/\.*" -not -path "*/docs/_build/*" -not -path "*/old/*" -not -path "*/build/*" -not -path "*/dist/*" | while read file; do
  echo "Fixing $file"
  autopep8 --in-place --aggressive --aggressive --max-line-length=100 "$file"
done

echo "Sorting imports with isort..."
# Fix import order with isort
find . -name "*.py" -not -path "*/\.*" -not -path "*/docs/_build/*" -not -path "*/old/*" -not -path "*/build/*" -not -path "*/dist/*" | while read file; do
  isort --profile black --line-length=100 "$file"
done

echo "Fixing end-of-file issues..."
# Ensure files end with a newline
find . -name "*.py" -not -path "*/\.*" -not -path "*/docs/_build/*" -not -path "*/old/*" -not -path "*/build/*" -not -path "*/dist/*" | while read file; do
  if [ "$(tail -c 1 "$file")" != "" ]; then
    echo "Adding newline to $file"
    echo "" >> "$file"
  fi
done

echo "Removing trailing whitespace..."
# Remove trailing whitespace
find . -name "*.py" -not -path "*/\.*" -not -path "*/docs/_build/*" -not -path "*/old/*" -not -path "*/build/*" -not -path "*/dist/*" | while read file; do
  sed -i '' -e 's/[ \t]*$//' "$file"
done

echo "Done! Some linting issues may still require manual fixes."
echo "Run './scripts/lint.sh' to check for remaining issues."