Quickstart Guide
==============

This guide will help you get a Mantl cluster up and running quickly using Vagrant for local development.

Prerequisites
------------

Before getting started, ensure you have the following installed:

* `Vagrant <https://www.vagrantup.com/downloads.html>`_ (version 1.8 or newer)
* `VirtualBox <https://www.virtualbox.org/wiki/Downloads>`_
* `Git <https://git-scm.com/downloads>`_
* `Ansible <https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html>`_
* `Terraform <https://www.terraform.io/downloads.html>`_

System requirements:

* At least 2GB of free RAM (4GB+ recommended)
* 20GB+ of free disk space
* 2+ CPU cores recommended

Step 1: Clone the Repository
---------------------------

First, clone the Mantl repository:

.. code-block:: bash

   git clone https://github.com/mantl/mantl.git
   cd mantl

Step 2: Start the Vagrant Environment
------------------------------------

Launch the Vagrant environment:

.. code-block:: bash

   vagrant up

This command will:

1. Download the required Vagrant box
2. Create and configure the virtual machines
3. Install all necessary components
4. Configure the Mantl cluster

The process can take 15-30 minutes depending on your hardware and internet connection.

Step 3: Access the Mantl UI
-------------------------

Once Vagrant has finished provisioning, you can access the Mantl UI:

1. Open a web browser
2. Go to https://192.168.100.101/
3. Accept the self-signed certificate warning
4. Log in with the default credentials:
   * Username: ``admin``
   * Password: The password is stored in the ``security.yml`` file. You can view it with:

   .. code-block:: bash

      grep 'nginx_admin_password:' security.yml | cut -d' ' -f2

Step 4: Explore the Components
----------------------------

From the Mantl UI, you can access various components:

* **Consul UI**: For service discovery and health monitoring
* **Mesos UI**: To view cluster resources and running tasks
* **Marathon UI**: For deploying and managing applications
* **Kubernetes Dashboard**: If Kubernetes is enabled

Step 5: Deploy Your First Application
-----------------------------------

Let's deploy a simple web application using Marathon:

1. In the Mantl UI, click on "Marathon"
2. Click "Create Application"
3. Input the following JSON configuration:

   .. code-block:: json

      {
        "id": "hello-world",
        "cpus": 0.1,
        "mem": 64,
        "instances": 1,
        "container": {
          "type": "DOCKER",
          "docker": {
            "image": "tutum/hello-world",
            "network": "BRIDGE",
            "portMappings": [
              { "containerPort": 80, "hostPort": 0 }
            ]
          }
        },
        "healthChecks": [
          {
            "protocol": "HTTP",
            "path": "/",
            "portIndex": 0,
            "gracePeriodSeconds": 30,
            "intervalSeconds": 10,
            "timeoutSeconds": 5,
            "maxConsecutiveFailures": 3
          }
        ]
      }

4. Click "Create Application"
5. Wait for the application to deploy
6. Once deployed, find the assigned port and access it via http://192.168.100.101:PORT/

Step 6: Clean Up
--------------

When you're done experimenting, you can stop or destroy the Vagrant environment:

To pause the VMs (preserving their state):

.. code-block:: bash

   vagrant suspend

To shut down the VMs (preserving their configuration):

.. code-block:: bash

   vagrant halt

To completely remove the VMs:

.. code-block:: bash

   vagrant destroy

Next Steps
---------

Now that you have a working Mantl environment, you can:

1. Explore the :doc:`components/index` documentation
2. Learn how to deploy on cloud providers with the :doc:`index` guides
3. Try the examples in the ``examples/`` directory:

   .. code-block:: bash

      cd examples/hello-world
      ./deploy.sh

4. Customize your configuration by editing ``terraform.tf`` and other files
5. Learn about security features in the :doc:`../security/index` section

Troubleshooting
--------------

If you encounter issues during the quickstart:

* Check the :doc:`../troubleshooting` guide
* Verify your system meets the minimum requirements
* Ensure all prerequisite software is installed correctly
* Try running ``vagrant provision`` if the initial setup didn't complete
* Check the Vagrant logs for specific errors

For more detailed instructions, see the full :doc:`vagrant` guide.