#
# Mantl • vSphere modern VM schema
# We clone from a template because hand‑crafting VMs is how horror stories begin.
# If this breaks, it was definitely written on a Friday afternoon (self‑deprecation quota met).
#
variable "use_modern_vsphere_schema" {
  description = "If true, create control nodes using modern vsphere_virtual_machine with clone/customization."
  default     = false
}

variable "customization_spec_name" {
  description = "Optional vSphere customization spec name to apply when cloning."
  default     = ""
}

# Look up vSphere inventory objects
data "vsphere_datacenter" "dc" {
  name = var.datacenter
}

data "vsphere_datastore" "ds" {
  name          = var.datastore
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_compute_cluster" "cluster" {
  name          = var.cluster
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_resource_pool" "pool" {
  name          = var.pool
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "net" {
  name          = var.network_label
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_virtual_machine" "template" {
  name          = var.template
  datacenter_id = data.vsphere_datacenter.dc.id
}

# Control nodes: because buttons need pushing and clusters need herding.
resource "vsphere_virtual_machine" "control_modern" {
  count            = var.use_modern_vsphere_schema ? var.control_count : 0
  name             = "${var.short_name}-control-${format("%02d", count.index + 1)}"
  folder           = var.folder
  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = var.control_cpu
  memory   = var.control_ram

  network_interface {
    network_id = data.vsphere_network.net.id
  }

  disk {
    label            = "disk0"
    size             = var.control_volume_size
    thin_provisioned = var.disk_type == "thin"
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id

    dynamic "customize" {
      for_each = var.customization_spec_name != "" ? [1] : []
      content {
        # Apply an existing customization spec by name
        # Note: this is a no-op block used to trigger spec application via the nested options below
        linux_options {
          host_name = "${var.short_name}-control-${format("%02d", count.index + 1)}"
          domain    = var.domain
        }
        network_interface {}
        dns_server_list = compact([var.dns_server1, var.dns_server2])
      }
    }
  }
}

# Worker nodes: where the actual work happens (don’t tell the control plane).
resource "vsphere_virtual_machine" "worker_modern" {
  count            = var.use_modern_vsphere_schema ? var.worker_count : 0
  name             = "${var.short_name}-worker-${format("%03d", count.index + 1)}"
  folder           = var.folder
  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = var.worker_cpu
  memory   = var.worker_ram

  network_interface {
    network_id = data.vsphere_network.net.id
  }

  disk {
    label            = "disk0"
    size             = var.worker_volume_size
    thin_provisioned = var.disk_type == "thin"
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id

    dynamic "customize" {
      for_each = var.customization_spec_name != "" ? [1] : []
      content {
        linux_options {
          host_name = "${var.short_name}-worker-${format("%03d", count.index + 1)}"
          domain    = var.domain
        }
        network_interface {}
        dns_server_list = compact([var.dns_server1, var.dns_server2])
      }
    }
  }
}

# Edge nodes: the bouncers of your cluster—IDs checked at the door.
resource "vsphere_virtual_machine" "edge_modern" {
  count            = var.use_modern_vsphere_schema ? var.edge_count : 0
  name             = "${var.short_name}-edge-${format("%02d", count.index + 1)}"
  folder           = var.folder
  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.ds.id

  num_cpus = var.edge_cpu
  memory   = var.edge_ram

  network_interface {
    network_id = data.vsphere_network.net.id
  }

  disk {
    label            = "disk0"
    size             = var.edge_volume_size
    thin_provisioned = var.disk_type == "thin"
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id

    dynamic "customize" {
      for_each = var.customization_spec_name != "" ? [1] : []
      content {
        linux_options {
          host_name = "${var.short_name}-edge-${format("%02d", count.index + 1)}"
          domain    = var.domain
        }
        network_interface {}
        dns_server_list = compact([var.dns_server1, var.dns_server2])
      }
    }
  }
}
