Testing Guide
=============

This document provides guidance on testing Mantl components.

Testing Principles
-----------------

* All new features should have tests
* All bug fixes should include a test that confirms the fix
* Tests should be automated when possible
* Integration tests should cover key workflows

Types of Tests
-------------

Unit Tests
^^^^^^^^^^

Unit tests verify individual functions and components in isolation. Mantl uses pytest for Python unit tests.

Example unit test::

    def test_parse_bool():
        from plugins.inventory.terraform import parse_bool
        
        assert parse_bool("true") is True
        assert parse_bool("True") is True
        assert parse_bool("false") is False
        assert parse_bool("False") is False
        
        with pytest.raises(ValueError):
            parse_bool("invalid")

Integration Tests
^^^^^^^^^^^^^^^^

Integration tests verify that multiple components work together correctly. These tests typically involve deploying and configuring services.

Mantl integration tests verify that:

* Services can be deployed successfully
* Services interact properly with each other
* Configuration changes apply correctly
* Upgrades work as expected

Example integration test workflow:

1. Deploy a minimal Mantl cluster
2. Verify all services are healthy
3. Deploy a sample application
4. Verify the application works correctly
5. Upgrade a component
6. Verify everything still works

End-to-End Tests
^^^^^^^^^^^^^^^

End-to-end tests verify complete workflows from a user's perspective. These tests often involve deploying a complete Mantl environment and testing real-world scenarios.

Example end-to-end test scenarios:

* Deploy Mantl on a cloud provider
* Launch services using Marathon
* Verify service discovery with Consul
* Test failover and recovery

Test Infrastructure
------------------

Continuous Integration
^^^^^^^^^^^^^^^^^^^^^

Mantl uses Travis CI for automated testing. The configuration is in ``.travis.yml``. 

Key CI workflows:

* Linting and syntax checking
* Unit tests
* Deployment tests on supported cloud providers

Local Testing
^^^^^^^^^^^^

For local development and testing:

1. Use Vagrant to create a local test environment::

    vagrant up

2. Run specific tests against your local environment::

    pytest tests/unit/path/to/test.py

3. Verify changes manually using the local environment

Testing Terraform Configurations
-------------------------------

For Terraform configurations:

1. Validate syntax::

    terraform validate

2. Format code consistently::

    terraform fmt

3. Run a plan to verify expected changes::

    terraform plan

4. For actual testing, apply the configuration to a test environment::

    terraform apply

Testing Ansible Roles
--------------------

For Ansible roles:

1. Use ansible-lint to check for best practices::

    ansible-lint roles/role_name

2. Test roles with Molecule when available::

    cd roles/role_name
    molecule test

3. Run playbooks with check mode to verify syntax and potentially catch issues::

    ansible-playbook playbook.yml --check

4. Run playbooks in a test environment to verify behavior

Testing Python Code
------------------

For Python code:

1. Run unit tests with pytest::

    pytest tests/unit/

2. Check code style with flake8::

    flake8 path/to/python/code

3. Measure test coverage::

    pytest --cov=mantl tests/

Test Documentation
-----------------

Document your tests:

* Include a README.md in each test directory explaining the tests
* Include setup instructions for running tests locally
* Document test scenarios and expected outcomes

Test Environments
----------------

Use appropriate test environments:

* Use Vagrant for local development and initial testing
* Use disposable cloud resources for integration and end-to-end tests
* Always clean up test environments after use

Additional Resources
------------------

* `pytest Documentation <https://docs.pytest.org/>`_
* `Molecule Documentation <https://molecule.readthedocs.io/>`_
* `Terraform Testing <https://www.terraform.io/docs/language/modules/testing-experiment.html>`_
* `Ansible Testing Strategies <https://docs.ansible.com/ansible/latest/reference_appendices/test_strategies.html>`_