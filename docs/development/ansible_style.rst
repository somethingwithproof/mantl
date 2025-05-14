Ansible Style Guide
=================

This document describes the coding style for Ansible roles and playbooks in Mantl.

General Guidelines
-----------------

* Use YAML files with ``.yml`` extension (not ``.yaml``)
* Use 2 spaces for indentation
* Limit lines to 100 characters when possible
* Follow `Ansible Best Practices <https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html>`_
* Always make tasks idempotent (running them multiple times has the same result as running once)

Directory Structure
-----------------

* Follow the standard Ansible directory structure for roles::

    roles/
      role_name/
        defaults/        # Default variables
          main.yml
        files/           # Files to be deployed as-is
        handlers/        # Handlers
          main.yml  
        meta/            # Role metadata
          main.yml
        tasks/           # Tasks
          main.yml
        templates/       # Jinja2 templates
        vars/            # Role variables
          main.yml

Naming Conventions
-----------------

* Use ``snake_case`` for variable names, role names, and task names
* Use descriptive names that indicate the purpose
* Prefix variables with the role name to avoid conflicts

Variables
---------

* Define default values in ``defaults/main.yml``
* Document all variables with comments
* Group related variables together
* Use consistent formatting for variable values

Example::

    # Number of Consul instances to create
    consul_node_count: 3
    
    # Consul version to install
    consul_version: "1.9.5"
    
    # Consul configuration options
    consul_config:
      data_dir: "/var/lib/consul"
      log_level: "INFO"
      encrypt: "{{ consul_encryption_key }}"

Tasks
-----

* Always name tasks clearly using action-oriented descriptions
* Use the YAML dictionary format for module arguments
* Use the ``state`` parameter explicitly
* Include a comment for complex tasks

Example::

    - name: Ensure Consul is installed
      package:
        name: consul
        state: present
        version: "{{ consul_version }}"
      
    - name: Create Consul configuration directory
      file:
        path: /etc/consul.d
        state: directory
        mode: '0750'
        owner: consul
        group: consul

Handlers
-------

* Use clear, specific handler names
* Name handlers based on the action they perform
* Use consistent naming patterns for related handlers

Example::

    - name: restart consul
      service:
        name: consul
        state: restarted
      
    - name: reload consul
      service:
        name: consul
        state: reloaded

Templates
--------

* Use ``.j2`` extension for Jinja2 templates
* Include a comment at the top of templates indicating it is managed by Ansible
* Use consistent indentation and spacing
* Add comments for complex sections

Example::

    # {{ ansible_managed }}
    
    [Unit]
    Description=Consul Agent
    After=network.target
    
    [Service]
    User={{ consul_user }}
    Group={{ consul_group }}
    ExecStart=/usr/bin/consul agent -config-dir=/etc/consul.d
    Restart=on-failure
    
    [Install]
    WantedBy=multi-user.target

Playbooks
--------

* Break complex playbooks into smaller, focused playbooks
* Use tags for optional tasks or groups of tasks
* Include clear comments and section separators for large playbooks
* Use become and become_user only when necessary

Example::

    ---
    - name: Configure Consul servers
      hosts: role=control
      become: true
      tags:
        - consul
        - configuration
      
      roles:
        - role: common
        - role: consul
          vars:
            consul_server: true
      
      tasks:
        - name: Verify Consul cluster health
          command: consul members
          register: consul_status
          changed_when: false
          check_mode: no
          run_once: true

Role Metadata
------------

* Include accurate metadata in ``meta/main.yml``
* Specify dependencies and compatible platforms
* Include appropriate tags for your role

Example::

    ---
    galaxy_info:
      author: Mantl Team
      description: Installs and configures Consul
      license: Apache-2.0
      min_ansible_version: 2.9
      platforms:
        - name: EL
          versions:
            - 7
            - 8
        - name: Ubuntu
          versions:
            - bionic
            - focal
      galaxy_tags:
        - mantl
        - consul
        - clustering
    
    dependencies:
      - role: common

Testing
-------

* Write tests for all roles using Molecule
* Test idempotence (running multiple times has the same result)
* Test different operating systems and configurations
* Include verification steps in tests

Additional Resources
------------------

* `Ansible Documentation <https://docs.ansible.com/ansible/latest/index.html>`_
* `Ansible Lint <https://github.com/ansible/ansible-lint>`_
* `Molecule <https://molecule.readthedocs.io/>`_ for testing roles