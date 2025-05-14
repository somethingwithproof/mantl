# Mantl Development Guide

This guide provides detailed information for developers working on Mantl.

## Development Environment

### Requirements

- Python 2.7 (most testing is done with this version)
- Terraform (version as specified in `.travis.yml`)
- Ansible (version as specified in `requirements.txt`)
- Vagrant 1.8+ and VirtualBox (for local testing)

### Setting Up

1. Make sure your development environment has all requirements installed
2. Fork and clone the repository
3. Install dependencies: `pip install -r requirements.txt`
4. Set up pre-commit hooks:

```bash
# Install pre-commit
pip install pre-commit

# Install the git hook scripts
pre-commit install

# (Optional) Run against all files
pre-commit run --all-files
```

The pre-commit hooks will automatically check your code for issues before each commit, helping maintain code quality and consistency.

## Code Quality

### Python Code

Python code in Mantl should follow [PEP 8](https://www.python.org/dev/peps/pep-0008/) guidelines. We recommend using a linter like `pylint` or `flake8` to ensure code quality.

Example configuration for `flake8` in your project:

```ini
[flake8]
max-line-length = 100
exclude = .git,__pycache__,docs/source/conf.py,old,build,dist
```

### Ansible Roles

Ansible roles should:
- Use YAML files with `.yml` extension
- Follow the [Ansible best practices](https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html)
- Include appropriate documentation in the `meta/main.yml` file
- Be idempotent (running multiple times has the same result as running once)

### Terraform Configurations

Terraform files should:
- Be formatted with `terraform fmt`
- Use consistent variable naming
- Contain appropriate documentation
- Be modular and reusable when possible

## Testing

### Types of Tests

1. **Unit Tests**: For testing individual functions and classes
2. **Integration Tests**: For testing interactions between components
3. **Acceptance Tests**: For testing user workflows

### Running Tests

- Unit tests can be run with `pytest tests/unit`
- Integration tests are typically run through the CI pipeline

## Documentation

Documentation for Mantl is built using Sphinx and reStructuredText.

### Documentation Structure

- `docs/`: Root directory for documentation
  - `getting_started/`: Instructions for getting started with Mantl
  - `components/`: Documentation for individual components
  - `security/`: Security-related documentation
  - `upgrading/`: Guides for upgrading Mantl

### Building Documentation

To build the documentation locally:

```
cd docs
make html
```

Then open `_build/html/index.html` in your browser.

### Documentation Style Guide

- Use sentence case for headings
- Be consistent with formatting
- Include code examples when relevant
- Keep language clear and concise

## Release Process

1. Update the version number in relevant files
2. Update the CHANGELOG.md with details about the release
3. Create a release branch
4. Create a tag for the release
5. Build and verify the release
6. Merge the release branch into master
7. Create a GitHub release with release notes

## Troubleshooting

### Common Issues

1. **Terraform state corruption**: If Terraform state becomes corrupted, try:
   ```
   terraform state pull > terraform.tfstate
   terraform state push terraform.tfstate
   ```

2. **Ansible connectivity issues**: Ensure SSH keys are properly configured and firewalls allow connections.

3. **Dependency conflicts**: Use virtual environments to isolate project dependencies.

### Debugging Tools

- Enable verbose logging in Ansible with `-v`, `-vv`, or `-vvv`
- Use `terraform plan -out=plan.out` to see what changes Terraform will make
- Inspect the Mantl API logs for API-related issues

## Getting Help

If you need assistance during development:

- Check the [FAQ](docs/faq.rst)
- Join the [Gitter chat](https://gitter.im/CiscoCloud/mantl)
- Open an issue on GitHub

## Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs/)
- [Ansible Documentation](https://docs.ansible.com/)
- [Mesos Documentation](http://mesos.apache.org/documentation/latest/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)