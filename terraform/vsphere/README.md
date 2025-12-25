# vSphere Terraform (mantl/terraform/vsphere)

Modernization in progress for Terraform 1.x.

Feature flags and new inputs:
- `use_modern_vsphere_schema` (bool) – enables modern resources defined in `modern.tf`.
- `customization_spec_name` – optional customization spec to apply (wire in your environment).
- Placeholders for datacenter/cluster/pool/network names are provided; supply values before enabling the flag.

- Removed legacy `remote-exec` provisioners.
- Added provider version constraints.
- A follow-up will add inputs for vSphere customization specs and align resource arguments with the current provider schema.
