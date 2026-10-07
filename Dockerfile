# SPDX-License-Identifier: Apache-2.0
# Official mise 2026.10.3 Debian image; tool versions come from mise.toml.
FROM ghcr.io/jdx/mise@sha256:58c4c847f5518a9a87a9a582886a485443dbca5eb024426577f58d586a0990c6

RUN sed -i 's|^URIs:.*deb.debian.org/|URIs: https://deb.debian.org/|' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends make=4.4.1-2 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 10001 -s /bin/bash mantl \
    && mkdir -p /home/mantl/.config/mise /workspace \
    && chown -R mantl:mantl /home/mantl /workspace

ENV MISE_DATA_DIR=/home/mantl/.local/share/mise \
    MISE_CONFIG_DIR=/home/mantl/.config/mise \
    MISE_CACHE_DIR=/home/mantl/.cache/mise \
    MISE_AUTO_INSTALL=false \
    MISE_EXEC_AUTO_INSTALL=false
ENV PATH="/home/mantl/.venv/bin:/home/mantl/.local/share/mise/shims:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

COPY --chown=mantl:mantl mise.toml /home/mantl/.config/mise/config.toml
COPY requirements.txt requirements-test.txt /opt/mantl-dependencies/

USER mantl
RUN mise install go python terraform kubectl kustomize helm uv \
    && mise exec -- python -m venv /home/mantl/.venv \
    && mise exec -- /home/mantl/.venv/bin/python -m pip install \
        --no-cache-dir --only-binary :all: --require-hashes \
        -r /opt/mantl-dependencies/requirements.txt \
        -r /opt/mantl-dependencies/requirements-test.txt

# Bind-mount the checkout here; workstation files are never copied into the image.
WORKDIR /workspace
CMD ["bash"]
