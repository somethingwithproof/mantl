# Contributing to Mantl

We are excited that you're interested in contributing to Mantl! This document provides a high-level overview of how you can get involved.

## Getting Started

1. **Fork the repository**: Start by forking the [Mantl repository](https://github.com/mantl/mantl).

2. **Clone your fork**: 
   ```
   git clone git@github.com:YOUR-USERNAME/mantl.git
   cd mantl
   ```

3. **Set up your environment**:
   - Install [Terraform](https://www.terraform.io/downloads.html) (required for development)
   - Install Python dependencies with `pip install -r requirements.txt`

4. **Add the upstream repository**:
   ```
   git remote add upstream git://github.com/mantl/mantl.git
   git fetch upstream
   ```

5. **Stay in sync**:
   ```
   git pull upstream master
   ```

## Development Workflow

1. **Create a branch**: Create a new branch for your feature or bugfix.
   ```
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**: Develop and test your changes locally.

3. **Follow code style**: 
   - Python code should follow [PEP 8](https://www.python.org/dev/peps/pep-0008/) guidelines
   - Ansible roles should follow [Ansible best practices](https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html)
   - Terraform configurations should be properly formatted with `terraform fmt`

4. **Write tests**: Add tests for your changes when applicable.

5. **Update documentation**: Update or add documentation as needed.

6. **Commit your changes**: Use clear commit messages that explain your changes.

7. **Push your branch**: Push your branch to your fork.
   ```
   git push origin feature/your-feature-name
   ```

8. **Submit a pull request**: Open a pull request against the upstream master branch.

## Pull Request Guidelines

* Keep PRs focused on a single topic.
* Provide a clear description of the changes.
* Link to related issues if applicable.
* Make sure tests pass.
* Update documentation as needed.

## Documentation

Documentation is written in reStructuredText and built using Sphinx. The documentation source is in the `docs` directory.

To build the documentation locally:

```
cd docs
make html
```

Then open `_build/html/index.html` in your browser.

## Testing

Before submitting a PR, make sure your changes pass all tests:

- Linting: Ensure your code passes style checks
- Unit tests: Run available tests
- Integration tests: When applicable, test your changes in a local environment

## Getting Help

If you have questions or need help, you can:

- Join the [Mantl chat room on Gitter](https://gitter.im/CiscoCloud/mantl)
- Open an issue on GitHub

## Code of Conduct

Please respect our [Code of Conduct](code-of-conduct.md) when participating in the Mantl community.

Thank you for your contributions!