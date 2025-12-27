# Ansible Collections for Mantl

This directory contains Ansible collection requirements for the Mantl platform.

## Prerequisites

- Ansible 13.x (ansible-core 2.17+)
- Python 3.9+

## Installation

Install all required collections:

```bash
ansible-galaxy collection install -r collections/requirements.yml
```

Install collections with upgrade:

```bash
ansible-galaxy collection install -r collections/requirements.yml --force
```

## Included Collections

### Cloud Providers
- **amazon.aws** - AWS modules and plugins
- **google.cloud** - Google Cloud Platform
- **azure.azcollection** - Microsoft Azure
- **oracle.oci** - Oracle Cloud Infrastructure
- **community.digitalocean** - DigitalOcean

### Kubernetes & Containers
- **kubernetes.core** - Kubernetes resource management
- **community.docker** - Docker container management

### Infrastructure & System
- **ansible.posix** - POSIX system utilities
- **ansible.utils** - General utilities and filters
- **community.general** - General purpose modules

### Networking
- **community.dns** - DNS management across providers

### Security
- **community.crypto** - Cryptographic operations
- **community.sops** - SOPS encrypted secrets
- **community.hashi_vault** - HashiCorp Vault integration

### Observability
- **community.grafana** - Grafana dashboards and datasources
- **prometheus.prometheus** - Prometheus monitoring

## Usage

After installation, collections can be used in playbooks:

```yaml
---
- name: Example playbook using collections
  hosts: localhost
  collections:
    - kubernetes.core
    - amazon.aws

  tasks:
    - name: Create Kubernetes deployment
      k8s:
        state: present
        definition: "{{ lookup('file', 'deployment.yml') }}"

    - name: Manage AWS EC2 instance
      ec2_instance:
        name: mantl-instance
        state: present
```

## Updating Collections

Check for updates:

```bash
ansible-galaxy collection list
```

Update all collections:

```bash
ansible-galaxy collection install -r collections/requirements.yml --force --upgrade
```

## Custom Collections Path

The `ansible.cfg` is configured to look for collections in:
1. `./collections` (project-local)
2. `~/.ansible/collections` (user-local)
3. `/usr/share/ansible/collections` (system-wide)

## Documentation

- [Ansible Collections Overview](https://docs.ansible.com/ansible/latest/user_guide/collections_using.html)
- [Ansible Galaxy](https://galaxy.ansible.com/)
