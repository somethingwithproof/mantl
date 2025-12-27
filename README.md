# Mantl 2026 — Cloud-Native Platform Engineering Toolkit

[![CI](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml/badge.svg)](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CNCF](https://img.shields.io/badge/CNCF-Graduated%20%26%20Incubating-blue)](https://www.cncf.io/projects/)

Mantl is a **Kubernetes-native, GitOps-driven, secure-by-default** platform engineering toolkit for building production-ready, multi-cloud infrastructure. Built entirely on CNCF Graduated and Incubating projects.

## 🌟 Overview

Deploy a complete, production-ready Kubernetes platform in minutes:

```bash
# Create cluster (any provider)
terraform -chdir=terraform/blueprints/aws-eks apply

# Bootstrap platform
kubectl apply -k clusters/production
```

**What you get:**
- ✅ GitOps continuous delivery (ArgoCD)
- ✅ eBPF networking with Gateway API (Cilium)
- ✅ Automated TLS certificates (cert-manager)
- ✅ Policy enforcement