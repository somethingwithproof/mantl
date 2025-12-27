#!/usr/bin/env python
"""
Dynamic inventory for Terraform - finds all `.tfstate` files below the working
directory and generates an inventory based on them.

This module implements Domain-Driven Design principles with clear separation
of concerns following SOLID principles.

Copyright 2015 Cisco Systems, Inc.
Licensed under the Apache License, Version 2.0
"""
from __future__ import annotations

import argparse
import json
import os
import re
from abc import ABC, abstractmethod
from collections import defaultdict
from dataclasses import dataclass, field
from functools import wraps
from pathlib import Path
from typing import Any, Callable, Dict, Iterator, List, Optional, Protocol, Tuple

VERSION = "0.4.0"


# Domain Models (DDD Entities and Value Objects)


@dataclass(frozen=True)
class IPAddress:
    """Value object representing an IP address."""

    address: str

    def is_private(self) -> bool:
        """Check if IP is in private range."""
        return self.address.startswith(("10.", "192.168.")) or self.address.startswith("172.")

    def __str__(self) -> str:
        return self.address


@dataclass
class HostAttributes:
    """Domain entity representing host attributes."""

    ansible_ssh_host: str
    ansible_ssh_port: int = 22
    ansible_ssh_user: str = "root"
    public_ipv4: str = ""
    private_ipv4: str = ""
    provider: str = ""
    role: str = "none"
    consul_dc: str = "none"
    consul_is_server: bool = False
    publicly_routable: bool = False
    ansible_python_interpreter: str = "python"
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert attributes to dictionary."""
        return {k: v for k, v in self.__dict__.items() if not k.startswith("_")}


@dataclass
class Host:
    """Domain entity representing a host."""

    name: str
    attributes: HostAttributes
    groups: List[str] = field(default_factory=list)

    def add_group(self, group: str) -> None:
        """Add a group to this host."""
        if group not in self.groups:
            self.groups.append(group)

    def add_mantl_groups(self) -> None:
        """Add Mantl-specific groups based on attributes."""
        self.add_group(f"role={self.attributes.role}")
        self.add_group(f"dc={self.attributes.consul_dc}")

        if self.attributes.publicly_routable:
            self.add_group("publicly_routable")


# Utilities


def clean_datacenter_name(dcname: str) -> str:
    """
    Clean datacenter name to meet Consul requirements.

    Consul DCs are strictly alphanumeric with underscores and hyphens.
    """
    return re.sub(r"[^\w_\-]", "-", dcname)


def parse_bool(string_form: str) -> bool:
    """Parse a string representation of a boolean."""
    token = string_form.lower()[0]

    if token == "t":
        return True
    if token == "f":
        return False

    raise ValueError(f"could not convert {string_form!r} to a bool")


class AttributeParser:
    """Utility class for parsing Terraform attributes."""

    @staticmethod
    def parse_prefix(source: Dict[str, Any], prefix: str, sep: str = ".") -> Iterator[Tuple[str, Any]]:
        """Parse attributes with a given prefix."""
        for compkey, value in source.items():
            try:
                curprefix, rest = compkey.split(sep, 1)
            except ValueError:
                continue

            if curprefix != prefix or rest == "#":
                continue

            yield rest, value

    @staticmethod
    def parse_attr_list(source: Dict[str, Any], prefix: str, sep: str = ".") -> List[Dict[str, Any]]:
        """Parse a list of attributes."""
        attrs: Dict[int, Dict[str, Any]] = defaultdict(dict)

        for compkey, value in AttributeParser.parse_prefix(source, prefix, sep):
            idx, key = compkey.split(sep, 1)
            attrs[int(idx)][key] = value

        return list(attrs.values())

    @staticmethod
    def parse_dict(source: Dict[str, Any], prefix: str, sep: str = ".") -> Dict[str, Any]:
        """Parse a dictionary of attributes."""
        return dict(AttributeParser.parse_prefix(source, prefix, sep))

    @staticmethod
    def parse_list(source: Dict[str, Any], prefix: str, sep: str = ".") -> List[Any]:
        """Parse a list from attributes."""
        return [value for _, value in AttributeParser.parse_prefix(source, prefix, sep)]


# Repository Pattern


class TerraformStateRepository:
    """Repository for accessing Terraform state files."""

    def __init__(self, root: Optional[Path] = None):
        self.root = root or Path.cwd()

    def find_state_files(self) -> Iterator[Path]:
        """Find all .tfstate files in the repository."""
        for dirpath, _, filenames in os.walk(self.root):
            for name in filenames:
                if Path(name).suffix == ".tfstate":
                    yield Path(dirpath) / name

    def load_resources(self) -> Iterator[Tuple[str, str, Dict[str, Any]]]:
        """Load all resources from state files."""
        for filename in self.find_state_files():
            with open(filename) as json_file:
                state = json.load(json_file)
                for module in state.get("modules", []):
                    module_name = module["path"][-1] if module["path"] else "root"
                    for key, resource in module.get("resources", {}).items():
                        yield module_name, key, resource


# Service Layer (Parser Protocol and Implementations)


class ResourceParser(Protocol):
    """Protocol defining the interface for resource parsers."""

    def parse(self, resource: Dict[str, Any], module_name: str) -> Host:
        """Parse a resource into a Host object."""
        ...


@dataclass
class BaseResourceParser(ABC):
    """Base class for resource parsers implementing common logic."""

    provider_name: str = ""

    @abstractmethod
    def extract_attributes(self, raw_attrs: Dict[str, Any], module_name: str) -> HostAttributes:
        """Extract host attributes from raw Terraform attributes."""
        ...

    @abstractmethod
    def extract_groups(self, attrs: HostAttributes, raw_attrs: Dict[str, Any]) -> List[str]:
        """Extract groups from attributes."""
        ...

    @abstractmethod
    def extract_name(self, raw_attrs: Dict[str, Any]) -> str:
        """Extract host name from raw attributes."""
        ...

    def apply_mantl_vars(self, attrs: HostAttributes) -> None:
        """Apply Mantl-specific variables."""
        attrs.consul_is_server = attrs.role == "control"

    def parse(self, resource: Dict[str, Any], module_name: str) -> Host:
        """Parse a resource into a Host object."""
        raw_attrs = resource["primary"]["attributes"]

        name = self.extract_name(raw_attrs)
        attributes = self.extract_attributes(raw_attrs, module_name)
        self.apply_mantl_vars(attributes)

        groups = self.extract_groups(attributes, raw_attrs)

        host = Host(name=name, attributes=attributes, groups=groups)
        host.add_mantl_groups()

        return host


class AWSInstanceParser(BaseResourceParser):
    """Parser for AWS EC2 instances."""

    provider_name = "aws"

    def extract_name(self, raw_attrs: Dict[str, Any]) -> str:
        return raw_attrs.get("tags.Name", raw_attrs["id"])

    def extract_attributes(self, raw_attrs: Dict[str, Any], module_name: str) -> HostAttributes:
        tags = AttributeParser.parse_dict(raw_attrs, "tags")

        attrs = HostAttributes(
            ansible_ssh_host=raw_attrs.get("public_ip", ""),
            ansible_ssh_port=22,
            ansible_ssh_user=tags.get("sshUser", "ec2-user"),
            public_ipv4=raw_attrs.get("public_ip", ""),
            private_ipv4=raw_attrs.get("private_ip", ""),
            provider=self.provider_name,
            role=tags.get("role", "none"),
            consul_dc=clean_datacenter_name(tags.get("dc", module_name)),
            ansible_python_interpreter=tags.get("python_bin", "python"),
            metadata=tags,
        )

        # Use private IP if sshPrivateIp tag is set
        if "sshPrivateIp" in tags:
            attrs.ansible_ssh_host = raw_attrs.get("private_ip", "")

        return attrs

    def extract_groups(self, attrs: HostAttributes, raw_attrs: Dict[str, Any]) -> List[str]:
        groups = [
            f"aws_ami={raw_attrs.get('ami', '')}",
            f"aws_az={raw_attrs.get('availability_zone', '')}",
            f"aws_key_name={raw_attrs.get('key_name', '')}",
            f"aws_tenancy={raw_attrs.get('tenancy', '')}",
        ]

        # Add tag groups
        for key, value in attrs.metadata.items():
            groups.append(f"aws_tag_{key}={value}")

        # Add VPC security groups
        vpc_sg_ids = AttributeParser.parse_list(raw_attrs, "vpc_security_group_ids")
        groups.extend(f"aws_vpc_security_group={group}" for group in vpc_sg_ids)

        return groups


class GCEInstanceParser(BaseResourceParser):
    """Parser for Google Compute Engine instances."""

    provider_name = "gce"

    def extract_name(self, raw_attrs: Dict[str, Any]) -> str:
        return raw_attrs["id"]

    def extract_attributes(self, raw_attrs: Dict[str, Any], module_name: str) -> HostAttributes:
        metadata = AttributeParser.parse_dict(raw_attrs, "metadata")
        interfaces = AttributeParser.parse_attr_list(raw_attrs, "network_interface")

        # Clean up nested interface data
        for interface in interfaces:
            interface["access_config"] = AttributeParser.parse_attr_list(interface, "access_config")
            interface = {k: v for k, v in interface.items() if "." not in k}

        attrs = HostAttributes(
            ansible_ssh_host="",
            ansible_ssh_port=22,
            ansible_ssh_user=metadata.get("ssh_user", "centos"),
            public_ipv4="",
            private_ipv4="",
            provider=self.provider_name,
            role=metadata.get("role", "none"),
            consul_dc=clean_datacenter_name(metadata.get("dc", module_name)),
            ansible_python_interpreter=metadata.get("python_bin", "python"),
            metadata=metadata,
            publicly_routable=False,
        )

        # Extract IPs from network interfaces
        try:
            nat_ip = (interfaces[0]["access_config"][0].get("nat_ip") or
                     interfaces[0]["access_config"][0].get("assigned_nat_ip"))
            if nat_ip:
                attrs.ansible_ssh_host = nat_ip
                attrs.public_ipv4 = nat_ip
                attrs.private_ipv4 = interfaces[0].get("address", "")
                attrs.publicly_routable = True
        except (KeyError, IndexError):
            pass

        return attrs

    def extract_groups(self, attrs: HostAttributes, raw_attrs: Dict[str, Any]) -> List[str]:
        groups = [
            f"gce_machine_type={raw_attrs.get('machine_type', '')}",
            f"gce_zone={raw_attrs.get('zone', '')}",
        ]

        # Add metadata groups (exclude sshKeys)
        for key, value in attrs.metadata.items():
            if key not in {"sshKeys"}:
                groups.append(f"gce_metadata_{key}={value}")

        # Add tags
        tags = AttributeParser.parse_list(raw_attrs, "tags")
        groups.extend(f"gce_tag={tag}" for tag in tags)

        # Add disk images
        disks = AttributeParser.parse_attr_list(raw_attrs, "disk")
        groups.extend(f"gce_image={disk.get('image', '')}" for disk in disks if disk.get("image"))

        if attrs.publicly_routable:
            groups.append("gce_publicly_routable")

        return groups


class DigitalOceanDropletParser(BaseResourceParser):
    """Parser for DigitalOcean droplets."""

    provider_name = "digitalocean"

    def extract_name(self, raw_attrs: Dict[str, Any]) -> str:
        return raw_attrs["name"]

    def extract_attributes(self, raw_attrs: Dict[str, Any], module_name: str) -> HostAttributes:
        metadata = json.loads(raw_attrs.get("user_data", "{}"))

        return HostAttributes(
            ansible_ssh_host=raw_attrs["ipv4_address"],
            ansible_ssh_port=22,
            ansible_ssh_user="root",
            public_ipv4=raw_attrs["ipv4_address"],
            private_ipv4=raw_attrs.get("ipv4_address_private", raw_attrs["ipv4_address"]),
            provider=self.provider_name,
            role=metadata.get("role", "none"),
            consul_dc=clean_datacenter_name(metadata.get("dc", raw_attrs["region"])),
            ansible_python_interpreter=metadata.get("python_bin", "python"),
            metadata=metadata,
        )

    def extract_groups(self, attrs: HostAttributes, raw_attrs: Dict[str, Any]) -> List[str]:
        return [
            f"do_image={raw_attrs.get('image', '')}",
            f"do_locked={parse_bool(raw_attrs.get('locked', 'false'))}",
            f"do_region={raw_attrs.get('region', '')}",
            f"do_size={raw_attrs.get('size', '')}",
            f"do_status={raw_attrs.get('status', '')}",
        ] + [f"do_metadata_{k}={v}" for k, v in attrs.metadata.items()]


class OpenStackInstanceParser(BaseResourceParser):
    """Parser for OpenStack compute instances."""

    provider_name = "openstack"

    def extract_name(self, raw_attrs: Dict[str, Any]) -> str:
        return raw_attrs["name"]

    def extract_attributes(self, raw_attrs: Dict[str, Any], module_name: str) -> HostAttributes:
        metadata = AttributeParser.parse_dict(raw_attrs, "metadata")

        attrs = HostAttributes(
            ansible_ssh_host=raw_attrs.get("access_ip_v4", ""),
            ansible_ssh_port=22,
            ansible_ssh_user=metadata.get("ssh_user", "centos"),
            public_ipv4=raw_attrs.get("access_ip_v4", ""),
            private_ipv4=raw_attrs.get("access_ip_v4", ""),
            provider=self.provider_name,
            role=metadata.get("role", "none"),
            consul_dc=clean_datacenter_name(metadata.get("dc", module_name)),
            ansible_python_interpreter=metadata.get("python_bin", "python"),
            metadata=metadata,
            publicly_routable=bool(raw_attrs.get("access_ip_v4")),
        )

        if "floating_ip" in raw_attrs:
            attrs.private_ipv4 = raw_attrs.get("network.0.fixed_ip_v4", "")

        return attrs

    def extract_groups(self, attrs: HostAttributes, raw_attrs: Dict[str, Any]) -> List[str]:
        flavor = AttributeParser.parse_dict(raw_attrs, "flavor", sep="_")
        image = AttributeParser.parse_dict(raw_attrs, "image", sep="_")

        groups = [
            f"os_image={image.get('name', '')}",
            f"os_flavor={flavor.get('name', '')}",
            f"os_region={raw_attrs.get('region', '')}",
        ]

        groups.extend(f"os_metadata_{k}={v}" for k, v in attrs.metadata.items())

        return groups


# Parser Registry (Strategy Pattern)


class ParserRegistry:
    """Registry for resource type parsers."""

    def __init__(self):
        self._parsers: Dict[str, BaseResourceParser] = {}

    def register(self, resource_type: str, parser: BaseResourceParser) -> None:
        """Register a parser for a resource type."""
        self._parsers[resource_type] = parser

    def get_parser(self, resource_type: str) -> Optional[BaseResourceParser]:
        """Get parser for a resource type."""
        return self._parsers.get(resource_type)

    def parse_resource(self, resource_type: str, resource: Dict[str, Any], module_name: str) -> Optional[Host]:
        """Parse a resource using the appropriate parser."""
        parser = self.get_parser(resource_type)
        if parser:
            return parser.parse(resource, module_name)
        return None


# Initialize parser registry
def create_parser_registry() -> ParserRegistry:
    """Create and populate the parser registry."""
    registry = ParserRegistry()

    registry.register("aws_instance", AWSInstanceParser())
    registry.register("google_compute_instance", GCEInstanceParser())
    registry.register("digitalocean_droplet", DigitalOceanDropletParser())
    registry.register("openstack_compute_instance_v2", OpenStackInstanceParser())

    return registry


# Application Services


class InventoryService:
    """Service for generating inventory from Terraform state."""

    def __init__(self, repository: TerraformStateRepository, parser_registry: ParserRegistry):
        self.repository = repository
        self.parser_registry = parser_registry

    def get_hosts(self) -> Iterator[Host]:
        """Get all hosts from Terraform state."""
        for module_name, key, resource in self.repository.load_resources():
            resource_type, _ = key.split(".", 1)
            host = self.parser_registry.parse_resource(resource_type, resource, module_name)
            if host:
                yield host

    def query_host(self, target: str) -> Dict[str, Any]:
        """Query a specific host."""
        for host in self.get_hosts():
            if host.name == target:
                return host.attributes.to_dict()
        return {}

    def query_list(self) -> Dict[str, Any]:
        """Query all hosts as a list."""
        groups: Dict[str, Dict[str, List[str]]] = defaultdict(lambda: {"hosts": []})
        meta: Dict[str, Dict[str, Any]] = {}

        for host in self.get_hosts():
            for group in set(host.groups):
                groups[group]["hosts"].append(host.name)
            meta[host.name] = host.attributes.to_dict()

        groups["_meta"] = {"hostvars": meta}
        return dict(groups)

    def query_hostfile(self) -> str:
        """Generate /etc/hosts snippet."""
        lines = ["## begin hosts generated by terraform.py ##"]

        for host in self.get_hosts():
            ssh_host = host.attributes.ansible_ssh_host.ljust(16)
            lines.append(f"{ssh_host}\t{host.name}")

        lines.append("## end hosts generated by terraform.py ##")
        return "\n".join(lines)


# CLI Interface


def main() -> None:
    """Main entry point for the inventory script."""
    parser = argparse.ArgumentParser(
        prog=__file__,
        description=__doc__,
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )

    modes = parser.add_mutually_exclusive_group(required=True)
    modes.add_argument("--list", action="store_true", help="list all variables")
    modes.add_argument("--host", help="list variables for a single host")
    modes.add_argument("--version", action="store_true", help="print version and exit")
    modes.add_argument("--hostfile", action="store_true", help="print hosts as a /etc/hosts snippet")

    parser.add_argument("--pretty", action="store_true", help="pretty-print output JSON")
    parser.add_argument("--nometa", action="store_true", help="with --list, exclude hostvars")

    default_root = os.environ.get(
        "TERRAFORM_STATE_ROOT",
        os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..")),
    )
    parser.add_argument("--root", default=default_root, help="custom root to search for `.tfstate`s in")

    args = parser.parse_args()

    if args.version:
        print(f"{__file__} {VERSION}")
        parser.exit()

    # Initialize services
    repository = TerraformStateRepository(Path(args.root))
    parser_registry = create_parser_registry()
    inventory_service = InventoryService(repository, parser_registry)

    # Execute requested query
    if args.list:
        output = inventory_service.query_list()
        if args.nometa:
            del output["_meta"]
        print(json.dumps(output, indent=4 if args.pretty else None))
    elif args.host:
        output = inventory_service.query_host(args.host)
        print(json.dumps(output, indent=4 if args.pretty else None))
    elif args.hostfile:
        output = inventory_service.query_hostfile()
        print(output)

    parser.exit()


if __name__ == "__main__":
    main()
