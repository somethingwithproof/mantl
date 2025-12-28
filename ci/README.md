# Mantl Scripts

This directory contains utility scripts for development, testing, and maintenance of Mantl.

## Available Scripts

### Linting

- **lint.sh** - Runs linting tools on the codebase to check code quality
  ```bash
  ./lint.sh
  ```
  This script runs:
  - flake8 for Python code quality
  - ansible-lint for Ansible roles and playbooks
  - terraform fmt checks for Terraform files

### Testing

- **test.sh** - Runs tests with coverage reporting
  ```bash
  ./test.sh
  ```
  This script:
  - Runs pytest with coverage
  - Generates HTML coverage report

## Adding New Scripts

When adding new scripts to this directory:

1. Make the script executable with `chmod +x script-name.sh`
2. Update this README.md to document the script
3. Follow the [shell style guide](https://google.github.io/styleguide/shellguide.html) for shell scripts
4. Include a comment header in the script describing its purpose
5. Add error handling and informative messages
6. Test the script thoroughly before committing

## Best Practices

- Ensure scripts have proper error handling
- Include usage information with `-h` or `--help` flags
- Use descriptive variable names
- Add comments explaining complex operations
- Keep scripts focused on a single task
- Use functions to organize code
- Follow security best practices (avoid hardcoded credentials, etc.)