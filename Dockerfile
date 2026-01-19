# Mantl dev tools image (modern)
# - Multi-stage, non-root runtime
# - Terraform pinned and checksum-verified
# - kubectl and kustomize included for CI validation

# Stage: terraform-downloader (downloads Terraform and verifies checksum)
FROM debian:bookworm-slim AS tfget
ARG TF_VERSION=1.14.3
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates unzip gnupg && rm -rf /var/lib/apt/lists/* \
 && arch=$(dpkg --print-architecture) \
 && case "$arch" in amd64) TF_ARCH=amd64 ;; arm64) TF_ARCH=arm64 ;; *) echo "unsupported arch: $arch" && exit 1 ;; esac \
 && cd /tmp \
 && curl -fsSLO https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_${TF_ARCH}.zip \
 && curl -fsSLO https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_SHA256SUMS \
 && grep "terraform_${TF_VERSION}_linux_${TF_ARCH}.zip" terraform_${TF_VERSION}_SHA256SUMS | sha256sum -c - \
 && unzip terraform_${TF_VERSION}_linux_${TF_ARCH}.zip -d /tmp/tf

# Final stage
FROM python:3.14-slim
ENV DEBIAN_FRONTEND=noninteractive
ARG KUBECTL_VERSION=1.30.4
ARG KUSTOMIZE_VERSION=5.4.2

RUN apt-get update && apt-get install -y --no-install-recommends \
      bash git ca-certificates curl unzip gnupg tar make \
    && rm -rf /var/lib/apt/lists/*

# Install kubectl
RUN arch=$(dpkg --print-architecture) \
 && curl -fsSLo /usr/local/bin/kubectl https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${arch}/kubectl \
 && chmod +x /usr/local/bin/kubectl

# Install kustomize
RUN arch=$(dpkg --print-architecture) \
 && curl -fsSLo /tmp/kustomize.tar.gz https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${arch}.tar.gz \
 && tar -xzf /tmp/kustomize.tar.gz -C /usr/local/bin kustomize \
 && rm -f /tmp/kustomize.tar.gz \
 && chmod +x /usr/local/bin/kustomize

# Install helm
ARG HELM_VERSION=3.15.3
RUN arch=$(dpkg --print-architecture) \
 && case "$arch" in amd64) HELM_ARCH=amd64 ;; arm64) HELM_ARCH=arm64 ;; *) echo "unsupported arch: $arch" && exit 1 ;; esac \
 && curl -fsSLo /tmp/helm.tgz https://get.helm.sh/helm-v${HELM_VERSION}-linux-${HELM_ARCH}.tar.gz \
 && tar -xzf /tmp/helm.tgz -C /tmp \
 && mv /tmp/linux-${HELM_ARCH}/helm /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm \
 && rm -rf /tmp/helm.tgz /tmp/linux-${HELM_ARCH}

# Terraform binary
COPY --from=tfget /tmp/tf/terraform /usr/local/bin/terraform

# Create non-root user
RUN useradd -m -u 10001 -s /bin/bash mantl
WORKDIR /workspace

# Python dependencies (install as root)
COPY requirements.txt requirements-test.txt ./
RUN pip install --no-cache-dir -U pip setuptools wheel \
 && pip install --no-cache-dir -r requirements.txt -r requirements-test.txt

# Project files
COPY --chown=mantl:mantl . .

# Drop to non-root for runtime
USER mantl

# Default command: drop into bash; CI will override with make/pytest/etc.
CMD ["bash"]
