# Mantl dev tools image (modern)
# - Multi-stage, non-root, no git clone at build time
# - Terraform pinned and checksum-verified
# - Python deps installed from repo context

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
FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
      bash python3 python3-pip git ca-certificates curl unzip \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 10001 -s /bin/bash mantl
USER mantl
WORKDIR /workspace

# Terraform binary
COPY --from=tfget /tmp/tf/terraform /usr/local/bin/terraform

# Python dependencies (two-step for better caching)
COPY --chown=mantl:mantl requirements.txt requirements-test.txt ./
RUN pip3 install --no-cache-dir -r requirements.txt -r requirements-test.txt

# Project files
COPY --chown=mantl:mantl . .

# Default command: drop into bash; CI will override with make/pytest/etc.
CMD ["bash"]
