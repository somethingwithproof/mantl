# Upgrading Mantl

This document provides guidelines for upgrading from one version of Mantl to another.

## General Upgrade Process

1. **Backup your configuration**: Before upgrading, always backup your configuration files, especially:
   - `terraform.tf`
   - `security.yml`
   - Any customized playbooks (e.g., `mantl.yml`)
   - Any custom Ansible inventory files

2. **Check for breaking changes**: Review the release notes and CHANGELOG for any breaking changes that might affect your deployment.

3. **Update the repository**:
   ```bash
   # If you cloned the repository
   git fetch --tags
   git checkout <new-version>

   # If you downloaded a release archive
   # Download the new version and extract it
   ```

4. **Run security-setup again**: Some versions require re-running the security-setup script to update certificates or security configuration.
   ```bash
   ./security-setup
   ```

5. **Run the upgrade playbook**:
   ```bash
   ansible-playbook -e @security.yml playbooks/upgrade-mantl.yml
   ```

## Version-Specific Upgrade Notes

### Upgrading to Latest Version

The latest version includes:
- Enhanced documentation
- Code quality improvements
- New development tools

Steps:
1. Follow the general upgrade process above
2. Review the new documentation in the `docs/` directory
3. Explore the new development tools in the `scripts/` directory

### Upgrading from 1.2.x to 1.3.x

Key changes:
- Updated component versions
- Infrastructure changes

Steps:
1. Follow the general upgrade process
2. Check for component version compatibility
3. Update any custom configurations

### Upgrading from 1.1.x to 1.2.x

Key changes:
- Security improvements
- Configuration structure changes

Steps:
1. Follow the general upgrade process
2. Re-run security-setup
3. Check for changes in configuration structure

## Post-Upgrade Steps

After upgrading:

1. **Verify the upgrade**: Check that all services are running correctly
   ```bash
   ansible-playbook playbooks/check-mantl.yml
   ```

2. **Update your documentation**: If you maintain internal documentation about your Mantl deployment, update it to reflect the new version.

3. **Test your applications**: Verify that your applications running on Mantl still function correctly.

## Troubleshooting Upgrades

If you encounter issues during the upgrade:

1. Check the Ansible output for specific errors
2. Consult the [Troubleshooting Guide](docs/troubleshooting.rst)
3. Look for specific issues in the GitHub repository
4. Reach out to the community on [Gitter](https://gitter.im/mantl/mantl)

## Rollback Procedure

If you need to roll back to a previous version:

1. Restore your backed-up configuration files
2. Check out the previous version tag
   ```bash
   git checkout <previous-version>
   ```
3. Re-run the deployment
   ```bash
   ansible-playbook -e @security.yml mantl.yml
   ```

## Component Version Compatibility

| Mantl Version | Consul | Vault | Mesos | Marathon | Kubernetes |
|---------------|--------|-------|-------|----------|------------|
| Latest        | X.Y.Z  | X.Y.Z | X.Y.Z | X.Y.Z    | X.Y.Z      |
| 1.2.0         | X.Y.Z  | X.Y.Z | X.Y.Z | X.Y.Z    | X.Y.Z      |
| 1.1.0         | X.Y.Z  | X.Y.Z | X.Y.Z | X.Y.Z    | X.Y.Z      |
| 1.0.3         | X.Y.Z  | X.Y.Z | X.Y.Z | X.Y.Z    | X.Y.Z      |

*Note: Replace X.Y.Z with the actual component versions for each Mantl release*