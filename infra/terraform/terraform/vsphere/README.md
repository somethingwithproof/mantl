# vSphere Terraform (mantl/terraform/vsphere)

Modernization (feature-flagged)
- use_modern_vsphere_schema (bool, default false): create control/worker/edge VMs using the modern vsphere_virtual_machine resource with template clone.
- customization_spec_name (string, optional): apply an existing vSphere customization spec during clone.

Inventory lookups
- modern.tf uses data sources (datacenter, datastore, resource pool, network, template) based on the existing input variables.

Example (control VMs via clone + optional customization)
```hcl
use_modern_vsphere_schema = true
customization_spec_name   = "mantl-control-spec"

# Typical required variables (examples)
datacenter     = "DC1"
cluster        = "Compute-Cluster"
pool           = "Compute-Cluster/Resources/Mantl"
datastore      = "vsanDatastore"
network_label  = "VM Network"
template       = "mantl-centos7-template"
short_name     = "mantl"
long_name      = "mantl"
control_count  = 3
control_cpu    = 2
control_ram    = 4096
control_volume_size = 20
ssh_user       = "centos"
ssh_key        = "~/.ssh/id_rsa.pub"
```

Notes
- Legacy resources remain the default path. When enabling the modern schema, set the corresponding legacy counts to 0 to avoid duplicate VMs.
- Further modernization will refactor legacy resources to the current provider schema.
