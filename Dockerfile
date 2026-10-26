# SPDX-License-Identifier: Apache-2.0
# Keep mise's binary update separate from the Debian development environment.
FROM ghcr.io/jdx/mise@sha256:0593528c637b0dae8a0acca66b5d49266cf262a811f06530d7273d3cc002f3a0 AS mise
FROM debian@sha256:4c25275446ca371766b8b093c696782eabce06263fccacb8d5d813b3dc44b308
COPY --from=mise /usr/local/bin/mise /usr/local/bin/mise

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git xz-utils unzip make=4.4.1-2 \
    && sed -i 's|^URIs:.*deb.debian.org/|URIs: https://deb.debian.org/|' /etc/apt/sources.list.d/debian.sources \
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
