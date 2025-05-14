Troubleshooting
==============

This guide provides solutions for common issues you might encounter when using Mantl.

Deployment Issues
----------------

Terraform Errors
^^^^^^^^^^^^^^^

**Issue**: Terraform fails with provider errors.

**Solution**:
   * Ensure you're using the correct Terraform version (check the README)
   * Update your provider plugins: ``terraform init -upgrade``
   * Check that your cloud credentials are properly configured
   * Review the error message for specific API or quota issues

**Issue**: Terraform state corruption.

**Solution**:
   * Back up your current state: ``cp terraform.tfstate terraform.tfstate.backup``
   * Try to fix state: ``terraform state pull > terraform.tfstate``
   * For severe corruption, you might need to start with a fresh state file

Ansible Errors
^^^^^^^^^^^^^

**Issue**: Ansible cannot connect to hosts.

**Solution**:
   * Verify SSH keys are properly configured
   * Check security groups allow SSH access (port 22)
   * Ensure hosts are reachable (ping test)
   * Try adding ``-vvv`` to the ansible-playbook command for verbose output

**Issue**: Ansible playbook fails on specific tasks.

**Solution**:
   * Check the error message for specific failures
   * Run the playbook with ``--start-at-task`` to resume after fixing the issue
   * Ensure all required variables are defined correctly
   * Verify that the target OS is supported by the role

Cluster Operation Issues
-----------------------

Consul Issues
^^^^^^^^^^^^

**Issue**: Consul nodes won't form a cluster.

**Solution**:
   * Check firewall rules - ensure ports 8300-8302 TCP/UDP are open between nodes
   * Verify consul is running: ``systemctl status consul``
   * Check logs: ``journalctl -u consul``
   * Ensure consistent Consul version across all nodes
   * Check for hostname/address resolution issues between nodes

**Issue**: Consul shows services as unhealthy.

**Solution**:
   * Check the specific health check output: ``consul members -detailed``
   * Verify the service is actually running
   * Check service configuration in ``/etc/consul/conf.d/``
   * Restart the affected service

Mesos Issues
^^^^^^^^^^^

**Issue**: Mesos agents not connecting to masters.

**Solution**:
   * Verify ZooKeeper is running and healthy
   * Check network connectivity between agents and masters
   * Examine logs: ``journalctl -u mesos-master`` or ``journalctl -u mesos-agent``
   * Ensure consistent Mesos version across the cluster
   * Check for proper DNS resolution of hostnames

**Issue**: Tasks fail to launch on Mesos.

**Solution**:
   * Check resource constraints (memory, CPU)
   * Verify docker is running properly on agent nodes
   * Examine framework-specific logs (e.g., Marathon)
   * Check for Docker image availability
   * Validate task definition for errors

Marathon Issues
^^^^^^^^^^^^^^

**Issue**: Applications fail to start in Marathon.

**Solution**:
   * Check the application configuration for errors
   * Verify container image availability
   * Examine task failure details in Marathon UI
   * Check for resource constraints on Mesos agents
   * Review Marathon logs: ``journalctl -u marathon``

**Issue**: Marathon leader election issues.

**Solution**:
   * Verify ZooKeeper is functioning correctly
   * Check Marathon logs on all control nodes
   * Restart Marathon service if necessary
   * Ensure consistent Marathon version across control nodes

Kubernetes Issues
^^^^^^^^^^^^^^^

**Issue**: Kubernetes pods stuck in pending state.

**Solution**:
   * Check node capacity: ``kubectl describe nodes``
   * Look for events: ``kubectl describe pod <pod-name>``
   * Verify networking plugins are functioning
   * Check for taints or affinity issues
   * Examine kubelet logs on agent nodes

**Issue**: Kubernetes service discovery not working.

**Solution**:
   * Check kube-dns/CoreDNS is running
   * Verify service definition is correct
   * Test DNS resolution from within a pod
   * Check network policies that might block traffic

Networking Issues
---------------

**Issue**: Services cannot communicate between nodes.

**Solution**:
   * Check security group rules for appropriate ports
   * Verify network connectivity between nodes
   * Ensure Consul DNS is functioning correctly
   * Test with simple tools like netcat or curl
   * Check for network overlay issues if using Calico

**Issue**: External access to services not working.

**Solution**:
   * Verify Traefik is running and configured correctly
   * Check that services are registered correctly in Consul
   * Confirm edge node security groups allow necessary traffic
   * Test from different network locations to identify scope of issue
   * Examine Traefik logs

Security Issues
-------------

**Issue**: TLS certificate errors.

**Solution**:
   * Check certificate expiration dates
   * Verify certificate paths in configurations
   * Ensure certificates are for the correct domains/IPs
   * Regenerate certificates if necessary with ``security-setup``

**Issue**: Authentication failures.

**Solution**:
   * Verify credentials are correct
   * Check for expired tokens or certificates
   * Ensure auth services (like Vault) are running
   * Check time synchronization across nodes

Diagnostic Commands
-----------------

System Diagnostics
^^^^^^^^^^^^^^^^^

.. code-block:: bash

   # Check system resources
   free -m               # Memory usage
   df -h                 # Disk usage
   top                   # CPU and process information
   
   # Check logs
   journalctl -u <service>          # Service-specific logs
   tail -f /var/log/syslog          # System logs
   
   # Network diagnostics
   netstat -tulpn                   # Open ports
   ping <host>                      # Basic connectivity
   traceroute <host>                # Network path
   tcpdump -i <interface> port 80   # Packet capture

Service Diagnostics
^^^^^^^^^^^^^^^^^

.. code-block:: bash

   # Consul
   consul members                   # List cluster members
   consul catalog services          # List registered services
   consul monitor                   # Watch Consul logs
   
   # Mesos
   mesos-cli state                  # Cluster state
   mesos-cli ps                     # Running tasks
   
   # Marathon
   marathon list                    # List applications
   
   # Kubernetes
   kubectl get nodes                # List nodes
   kubectl get pods --all-namespaces # List all pods
   kubectl describe pod <pod>       # Detailed pod info
   kubectl logs <pod>               # Pod logs

Getting Help
-----------

If you're unable to resolve an issue using this guide:

1. Check the `GitHub issues <https://github.com/mantl/mantl/issues>`_ to see if it's a known problem
2. Join the `Gitter chat <https://gitter.im/mantl/mantl>`_ for community support
3. Open a new issue on GitHub with detailed information about your problem
4. For commercial support, contact the maintaining organizations