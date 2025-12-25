# Modern vSphere VMs (skeleton behind feature flag)

# NOTE: This is a schema-aligned skeleton for current vsphere provider.
# It is intentionally minimal and disabled by default via use_modern_vsphere_schema.

# Data sources (provide IDs via variables or discover by name)
variable "datacenter_name" { type = string, default = "" }
variable "cluster_name"    { type = string, default = "" }
variable "resource_pool_name" { type = string, default = "" }
variable "network_name"    { type = string, default = "" }

data "vsphere_datacenter" "dc" {
  count = var.use_modern_vsphere_schema && var.datacenter_name != "" ? 1 : 0
  name  = var.datacenter_name
}

data "vsphere_compute_cluster" "cluster" {
  count         = var.use_modern_vsphere_schema && var.cluster_name != "" ? 1 : 0
  name          = var.cluster_name
  datacenter_id = data.vsphere_datacenter.dc[0].id
}

data "vsphere_resource_pool" "pool" {
  count         = var.use_modern_vsphere_schema && var.resource_pool_name != "" ? 1 : 0
  name          = var.resource_pool_name
  datacenter_id = data.vsphere_datacenter.dc[0].id
}

data "vsphere_network" "net" {
  count         = var.use_modern_vsphere_schema && var.network_name != "" ? 1 : 0
  name          = var.network_name
  datacenter_id = data.vsphere_datacenter.dc[0].id
}

# Control (modern)
resource "vsphere_virtual_machine" "control_modern" {
  count            = var.use_modern_vsphere_schema ? var.control_count : 0
  name             = format("%s-control-%02d", var.short_name, count.index + 1)
  resource_pool_id = try(data.vsphere_resource_pool.pool[0].id, null)
  datastore_id     = null # set via clone block/template defaults or add variable

  num_cpus = var.control_cpu
  memory   = var.control_ram

  network_interface {
    network_id = try(data.vsphere_network.net[0].id, null)
  }

  disk {
    label = "disk0"
    size  = var.control_volume_size
  }

  # Clone/template usage to be wired with your template variable
  # clone {
  #   template_uuid = var.template
  #   customize {
  #     linux_options {
  #       host_name = format("%s-control-%02d", var.short_name, count.index + 1)
  #       domain    = var.domain
  #     }
  #     network_interface {}
  #     dns_server_list = compact([var.dns_server1, var.dns_server2])
  #   }
  # }
}
