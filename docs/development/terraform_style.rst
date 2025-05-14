Terraform Style Guide
===================

This document describes the coding style for Terraform configurations in Mantl.

General Guidelines
-----------------

* Format all Terraform code using ``terraform fmt``
* Use clear, descriptive variable and resource names
* Comment complex sections of code
* Organize resources in logical groups

File Structure
--------------

* Use separate files for different types of resources or logical groupings
* Common file names include:
  
  * ``main.tf`` - Primary resources
  * ``variables.tf`` - Input variables
  * ``outputs.tf`` - Output variables
  * ``versions.tf`` - Required provider versions

Naming Conventions
-----------------

* Use ``snake_case`` for resource names, variable names, and outputs
* Use descriptive names that indicate the purpose of the resource
* Prefix resource names with their type (e.g., ``aws_instance_control_node``)

Variables
---------

* Include a description for all variables
* Specify a type for each variable
* Provide default values when appropriate
* Group related variables together

Example::

    variable "control_node_count" {
      description = "Number of control nodes to create"
      type        = number
      default     = 3
    }
    
    variable "control_node_size" {
      description = "Instance size for control nodes"
      type        = string
      default     = "t3.large"
    }

Resources
---------

* Group related resources together
* Use consistent and descriptive resource names
* Use resource arguments in the standard order shown in the Terraform documentation
* Add tags to all resources that support them

Example::

    resource "aws_instance" "control_node" {
      count         = var.control_node_count
      ami           = var.control_node_ami
      instance_type = var.control_node_size
      subnet_id     = var.subnet_id
      
      tags = {
        Name  = "mantl-control-${count.index}"
        Role  = "control"
        Owner = "mantl"
      }
    }

Modules
-------

* Create reusable modules for common infrastructure patterns
* Include a ``README.md`` with example usage in each module
* Use consistent argument ordering in module blocks
* Include descriptions for all module inputs and outputs

Outputs
-------

* Provide outputs for important resource attributes
* Include a description for each output
* Use consistent naming formats for outputs

Example::

    output "control_node_ips" {
      description = "Public IP addresses of control nodes"
      value       = aws_instance.control_node[*].public_ip
    }

Version Constraints
------------------

* Specify required provider versions to ensure compatibility
* Use ``~>`` operator for allowing minor version updates but restricting major version changes

Example::

    terraform {
      required_version = "~> 0.14.0"
      
      required_providers {
        aws = {
          source  = "hashicorp/aws"
          version = "~> 3.0"
        }
      }
    }

State Management
--------------

* Use remote state storage for production environments
* Document the state backend configuration
* Avoid storing sensitive data in state files

Additional Resources
------------------

* `Terraform Style Conventions <https://www.terraform.io/docs/language/syntax/style.html>`_
* `Standard Module Structure <https://www.terraform.io/docs/language/modules/develop/structure.html>`_