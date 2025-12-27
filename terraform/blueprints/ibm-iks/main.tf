terraform {
  required_version = ">= 1.10.0"
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = ">= 1.65.0"
    }
  }
}

provider "ibm" {
  ibmcloud_api_key = var.ibmcloud_api_key
  region           = var.region
}

# Get resource group
data "ibm_resource_group" "resource_group" {
  name = var.resource_group
}

# Get available zones in region
data "ibm_is_zones" "zones" {
  region = var.region
}

# Create VPC
resource "ibm_is_vpc" "iks_vpc" {
  name           = "${var.cluster_name}-vpc"
  resource_group = data.ibm_resource_group.resource_group.id
}

# Create Public Gateway for each zone
resource "ibm_is_public_gateway" "pgw" {
  count          = var.zones_count
  name           = "${var.cluster_name}-pgw-${count.index + 1}"
  vpc            = ibm_is_vpc.iks_vpc.id
  zone           = data.ibm_is_zones.zones.zones[count.index]
  resource_group = data.ibm_resource_group.resource_group.id
}

# Create subnet for each zone
resource "ibm_is_subnet" "iks_subnet" {
  count                    = var.zones_count
  name                     = "${var.cluster_name}-subnet-${count.index + 1}"
  vpc                      = ibm_is_vpc.iks_vpc.id
  zone                     = data.ibm_is_zones.zones.zones[count.index]
  total_ipv4_address_count = 256
  public_gateway           = ibm_is_public_gateway.pgw[count.index].id
  resource_group           = data.ibm_resource_group.resource_group.id
}

# Create IKS cluster
resource "ibm_container_vpc_cluster" "iks_cluster" {
  name              = var.cluster_name
  vpc_id            = ibm_is_vpc.iks_vpc.id
  flavor            = var.worker_flavor
  worker_count      = var.workers_per_zone
  kube_version      = var.kubernetes_version
  resource_group_id = data.ibm_resource_group.resource_group.id
  wait_till         = "MasterNodeReady"

  dynamic "zones" {
    for_each = range(var.zones_count)
    content {
      subnet_id = ibm_is_subnet.iks_subnet[zones.value].id
      name      = data.ibm_is_zones.zones.zones[zones.value]
    }
  }

  disable_public_service_endpoint = var.disable_public_endpoint
  entitlement                     = var.entitlement

  tags = concat(
    var.tags,
    ["mantl-platform"]
  )
}

# Create worker pool (optional additional pool)
resource "ibm_container_vpc_worker_pool" "additional_pool" {
  count             = var.create_additional_pool ? 1 : 0
  cluster           = ibm_container_vpc_cluster.iks_cluster.id
  worker_pool_name  = "${var.cluster_name}-pool-2"
  flavor            = var.additional_pool_flavor
  vpc_id            = ibm_is_vpc.iks_vpc.id
  worker_count      = var.additional_pool_workers_per_zone
  resource_group_id = data.ibm_resource_group.resource_group.id

  dynamic "zones" {
    for_each = range(var.zones_count)
    content {
      subnet_id = ibm_is_subnet.iks_subnet[zones.value].id
      name      = data.ibm_is_zones.zones.zones[zones.value]
    }
  }
}

# Configure kubectl access
resource "null_resource" "configure_kubectl" {
  depends_on = [ibm_container_vpc_cluster.iks_cluster]

  provisioner "local-exec" {
    command = <<-EOT
      ibmcloud login --apikey ${var.ibmcloud_api_key} -r ${var.region} -g ${var.resource_group}
      ibmcloud ks cluster config --cluster ${ibm_container_vpc_cluster.iks_cluster.id} --admin
    EOT
  }
}
