<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | 3.3.0 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | 3.3.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_helm"></a> [helm](#provider\_helm) | 3.3.0 |
| <a name="provider_kubernetes"></a> [kubernetes](#provider\_kubernetes) | 3.3.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [helm_release.cert_manager](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [helm_release.cilium](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [helm_release.external_secrets](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [helm_release.falco](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [helm_release.sealed_secrets](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [helm_release.tailscale_operator](https://registry.terraform.io/providers/hashicorp/helm/3.3.0/docs/resources/release) | resource |
| [kubernetes_config_map.security_dashboard](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/config_map) | resource |
| [kubernetes_namespace.cert_manager](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_namespace.cilium](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_namespace.external_secrets](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_namespace.falco](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_namespace.sealed_secrets](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_namespace.tailscale](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/namespace) | resource |
| [kubernetes_secret.tailscale_auth](https://registry.terraform.io/providers/hashicorp/kubernetes/3.3.0/docs/resources/secret) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_enable_cert_manager"></a> [enable\_cert\_manager](#input\_enable\_cert\_manager) | Enable cert-manager for automated TLS certificate management | `bool` | `true` | no |
| <a name="input_enable_cilium"></a> [enable\_cilium](#input\_enable\_cilium) | Enable Cilium CNI with Hubble for network observability (flow logs alternative) | `bool` | `false` | no |
| <a name="input_enable_external_secrets"></a> [enable\_external\_secrets](#input\_enable\_external\_secrets) | Enable External Secrets Operator for external secrets management | `bool` | `false` | no |
| <a name="input_enable_falco"></a> [enable\_falco](#input\_enable\_falco) | Enable Falco for runtime security monitoring | `bool` | `false` | no |
| <a name="input_enable_flow_logs_export"></a> [enable\_flow\_logs\_export](#input\_enable\_flow\_logs\_export) | Enable export of Cilium flow logs to external storage | `bool` | `true` | no |
| <a name="input_enable_sealed_secrets"></a> [enable\_sealed\_secrets](#input\_enable\_sealed\_secrets) | Enable Sealed Secrets for encrypted secrets in Git (KMS alternative) | `bool` | `false` | no |
| <a name="input_enable_security_dashboard"></a> [enable\_security\_dashboard](#input\_enable\_security\_dashboard) | Enable Grafana dashboard for security monitoring | `bool` | `true` | no |
| <a name="input_enable_tailscale"></a> [enable\_tailscale](#input\_enable\_tailscale) | Enable Tailscale VPN for secure cluster access (private endpoint alternative) | `bool` | `false` | no |
| <a name="input_falco_ui_enabled"></a> [falco\_ui\_enabled](#input\_falco\_ui\_enabled) | Enable Falco UI for alert visualization | `bool` | `false` | no |
| <a name="input_falco_webhook_url"></a> [falco\_webhook\_url](#input\_falco\_webhook\_url) | Webhook URL for Falco alerts (e.g., Slack, PagerDuty) | `string` | `""` | no |
| <a name="input_flow_logs_s3_bucket"></a> [flow\_logs\_s3\_bucket](#input\_flow\_logs\_s3\_bucket) | S3-compatible bucket for flow logs export (leave empty to disable) | `string` | `""` | no |
| <a name="input_hubble_ui_hosts"></a> [hubble\_ui\_hosts](#input\_hubble\_ui\_hosts) | Hostnames for Hubble UI ingress | `list(string)` | `[]` | no |
| <a name="input_hubble_ui_ingress_enabled"></a> [hubble\_ui\_ingress\_enabled](#input\_hubble\_ui\_ingress\_enabled) | Enable ingress for Hubble UI | `bool` | `false` | no |
| <a name="input_tailscale_auth_key"></a> [tailscale\_auth\_key](#input\_tailscale\_auth\_key) | Tailscale auth key (deprecated - use OAuth) | `string` | `""` | no |
| <a name="input_tailscale_client_id"></a> [tailscale\_client\_id](#input\_tailscale\_client\_id) | Tailscale OAuth client ID | `string` | `""` | no |
| <a name="input_tailscale_client_secret"></a> [tailscale\_client\_secret](#input\_tailscale\_client\_secret) | Tailscale OAuth client secret | `string` | `""` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_cilium_enabled"></a> [cilium\_enabled](#output\_cilium\_enabled) | Whether Cilium is enabled |
| <a name="output_external_secrets_enabled"></a> [external\_secrets\_enabled](#output\_external\_secrets\_enabled) | Whether External Secrets Operator is enabled |
| <a name="output_falco_enabled"></a> [falco\_enabled](#output\_falco\_enabled) | Whether Falco is enabled |
| <a name="output_hubble_ui_url"></a> [hubble\_ui\_url](#output\_hubble\_ui\_url) | Hubble UI URL (if ingress enabled) |
| <a name="output_sealed_secrets_enabled"></a> [sealed\_secrets\_enabled](#output\_sealed\_secrets\_enabled) | Whether Sealed Secrets is enabled |
| <a name="output_security_features_summary"></a> [security\_features\_summary](#output\_security\_features\_summary) | Summary of enabled security features |
| <a name="output_tailscale_enabled"></a> [tailscale\_enabled](#output\_tailscale\_enabled) | Whether Tailscale VPN is enabled |
<!-- END_TF_DOCS -->
