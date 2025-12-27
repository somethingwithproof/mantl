terraform {
  required_version = ">= 1.10.0"
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = ">= 5.0"
    }
  }
}

provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

# Get availability domains
data "oci_identity_availability_domains" "ads" {
  compartment_id = var.compartment_id
}

# VCN for OKE cluster
resource "oci_core_vcn" "oke_vcn" {
  compartment_id = var.compartment_id
  display_name   = "${var.cluster_name}-vcn"
  cidr_blocks    = ["10.0.0.0/16"]
  dns_label      = replace(var.cluster_name, "-", "")
}

# Internet Gateway
resource "oci_core_internet_gateway" "oke_igw" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${var.cluster_name}-igw"
  enabled        = true
}

# Route table for public subnet
resource "oci_core_route_table" "oke_route_table" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${var.cluster_name}-routetable"

  route_rules {
    network_entity_id = oci_core_internet_gateway.oke_igw.id
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
  }
}

# Security list
resource "oci_core_security_list" "oke_security_list" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${var.cluster_name}-seclist"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 6443
      max = 6443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "10.0.0.0/16"
    tcp_options {
      min = 1
      max = 65535
    }
  }
}

# Subnet for worker nodes
resource "oci_core_subnet" "oke_node_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = "10.0.10.0/24"
  display_name      = "${var.cluster_name}-node-subnet"
  route_table_id    = oci_core_route_table.oke_route_table.id
  security_list_ids = [oci_core_security_list.oke_security_list.id]
  dns_label         = "nodes"
}

# Subnet for load balancers
resource "oci_core_subnet" "oke_lb_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = "10.0.20.0/24"
  display_name      = "${var.cluster_name}-lb-subnet"
  route_table_id    = oci_core_route_table.oke_route_table.id
  security_list_ids = [oci_core_security_list.oke_security_list.id]
  dns_label         = "loadbalancers"
}

# Subnet for API endpoint
resource "oci_core_subnet" "oke_api_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = "10.0.0.0/28"
  display_name      = "${var.cluster_name}-api-subnet"
  route_table_id    = oci_core_route_table.oke_route_table.id
  security_list_ids = [oci_core_security_list.oke_security_list.id]
  dns_label         = "api"
}

# OKE Cluster
resource "oci_containerengine_cluster" "oke_cluster" {
  compartment_id     = var.compartment_id
  kubernetes_version = var.kubernetes_version
  name               = var.cluster_name
  vcn_id             = oci_core_vcn.oke_vcn.id

  endpoint_config {
    is_public_ip_enabled = true
    subnet_id            = oci_core_subnet.oke_api_subnet.id
  }

  options {
    service_lb_subnet_ids = [oci_core_subnet.oke_lb_subnet.id]

    add_ons {
      is_kubernetes_dashboard_enabled = false
      is_tiller_enabled               = false
    }

    kubernetes_network_config {
      pods_cidr     = "10.244.0.0/16"
      services_cidr = "10.96.0.0/16"
    }
  }
}

# Node Pool
resource "oci_containerengine_node_pool" "oke_node_pool" {
  cluster_id         = oci_containerengine_cluster.oke_cluster.id
  compartment_id     = var.compartment_id
  kubernetes_version = var.kubernetes_version
  name               = "${var.cluster_name}-pool"

  node_config_details {
    placement_configs {
      availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
      subnet_id           = oci_core_subnet.oke_node_subnet.id
    }
    placement_configs {
      availability_domain = data.oci_identity_availability_domains.ads.availability_domains[1].name
      subnet_id           = oci_core_subnet.oke_node_subnet.id
    }
    placement_configs {
      availability_domain = data.oci_identity_availability_domains.ads.availability_domains[2].name
      subnet_id           = oci_core_subnet.oke_node_subnet.id
    }
    size = var.node_count
  }

  node_shape = var.node_shape

  node_shape_config {
    memory_in_gbs = var.node_memory_gb
    ocpus         = var.node_ocpus
  }

  node_source_details {
    image_id    = var.node_image_id
    source_type = "IMAGE"
  }

  initial_node_labels {
    key   = "name"
    value = var.cluster_name
  }
}

# Configure kubectl access
resource "null_resource" "configure_kubectl" {
  depends_on = [oci_containerengine_cluster.oke_cluster]

  provisioner "local-exec" {
    command = <<-EOT
      oci ce cluster create-kubeconfig \
        --cluster-id ${oci_containerengine_cluster.oke_cluster.id} \
        --file $HOME/.kube/config \
        --region ${var.region} \
        --token-version 2.0.0 \
        --kube-endpoint PUBLIC_ENDPOINT
    EOT
  }
}
