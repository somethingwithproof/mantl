FAQs
====

General Questions
----------------

What is the relationship between Mantl and `OpenStack Magnum <https://wiki.openstack.org/wiki/Magnum>`_?
---------------------------------------------------------------------------------------

Mantl and Magnum are currently not integrated. However, the projects could
complement one another. Magnum provides an OpenStack API to instantiate a
containerized environment within an OpenStack cloud. Magnum supports a range
of container clustering implementations and Operating System distributions.
Please refer to the `Magnum wiki <https://wiki.openstack.org/wiki/Magnum>`_
for additional Magnum details.

Mantl is an end-to-end solution for deploying and managing a microservices
infrastructure. Mantl hosts are provisioned to OpenStack and other supported
environments using `Terraform <https://www.terraform.io/>`_. Terraform
configuration files manage OpenStack services such as compute,
block storage, networking, etc. required to instantiate a Mantl host
to an OpenStack cloud. The Terraform `OpenStack Provider
<https://www.terraform.io/docs/providers/openstack/index.html>`_ would need to be
updated since it does not support Magnum. If/when this is accomplished, adding
Magnum support to Mantl should be straightforward.

Can I use Mantl with `Kubernetes <http://kubernetes.io>`_?
----------------------------------------------------------------

Kubernetes is an open source orchestration system for Docker containers.
It handles scheduling onto nodes in a compute cluster and actively manages
workloads to ensure that their state matches the users declared intentions.
Using the concepts of "labels" and "pods", it groups the containers which
make up an application into logical units for management and discovery.

Mantl has integrated both Apache Mesos and Kubernetes into it's container stack.
This integration provides users the freedom to choose the best scheduler for their
workloads promoting greater flexibility and choice.

Containers are great for running stateless applications but what about data/stateful services?
------------------------------------------------------------------------------------------------

The container ecosystem is moving quickly, and durable persistent storage is one area
that has received consistent attention. Mantl currently supports GlusterFS as an
`addon <http://docs.mantl.io/en/latest/components/glusterfs.html>`_ for shared
persistent storage. Even without this software, there are databases and patterns that
can provide reliable and consistent data for various use cases. For example, it is 
possible to run MongoDB, Redis, or Cassandra in a way that provides a consistent distributed quorum.

Development Questions
--------------------

Where can I find documentation about contributing to Mantl?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Mantl now has comprehensive documentation for contributors:

- `CONTRIBUTING.md <https://github.com/mantl/mantl/blob/master/CONTRIBUTING.md>`_ - Basic guidelines for contributing to Mantl
- `DEVELOPMENT.md <https://github.com/mantl/mantl/blob/master/DEVELOPMENT.md>`_ - Detailed development guide
- :doc:`development/index` - Full development documentation

How do I report a bug or request a feature?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Please open an issue on the `Mantl GitHub repository <https://github.com/mantl/mantl/issues>`_. 
Be sure to provide:

1. Clear steps to reproduce the issue (for bugs)
2. Expected behavior and actual behavior
3. Environment details (OS, cloud provider, etc.)

For feature requests, describe the feature and why it would be valuable to Mantl users.

How can I check if my code follows Mantl's style guidelines?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Mantl includes several tools to check code quality:

1. Run ``make lint`` to check all code
2. Run ``./scripts/lint.sh`` for more detailed linting
3. Use ``flake8`` for Python code
4. Use ``ansible-lint`` for Ansible roles and playbooks
5. Use ``terraform fmt`` for Terraform files

Documentation Questions
---------------------

How do I build the documentation locally?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

To build the documentation, run:

.. code-block:: bash

   # Install dependencies
   pip install -r requirements.txt
   
   # Build the docs
   cd docs
   make html
   
   # View the docs
   open _build/html/index.html

How can I contribute to the documentation?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Documentation improvements are always welcome! To contribute:

1. Fork the repository
2. Make your changes to files in the ``docs/`` directory
3. Build the documentation locally to verify your changes
4. Submit a pull request

For more details, see :doc:`development/index`.

