# Example tfvars for vSphere modern paths
use_modern_vsphere_schema = true
customization_spec_name   = "mantl-control-spec"

# Inventory
datacenter    = "DC1"
cluster       = "Compute-Cluster"
pool          = "Compute-Cluster/Resources/Mantl"
datastore     = "vsanDatastore"
network_label = "VM Network"
template      = "mantl-centos7-template"

# VM settings
short_name = "mantl"
long_name  = "mantl"
ssh_user   = "centos"
ssh_key    = "~/.ssh/id_rsa.pub"

control_count       = 3
control_cpu         = 2
control_ram         = 4096
control_volume_size = 20

worker_count       = 2
worker_cpu         = 2
worker_ram         = 4096
worker_volume_size = 20

edge_count       = 2
edge_cpu         = 2
edge_ram         = 4096
edge_volume_size = 20
