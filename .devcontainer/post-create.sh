#!/bin/bash
set -e

echo "🚀 Setting up Mantl development environment..."

# Install additional CLI tools
echo "📦 Installing additional tools..."

# Install kind (Kubernetes in Docker)
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install kubectl krew (plugin manager)
(
  cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"

# Install useful kubectl plugins
kubectl krew install ctx ns stern tree whoami

# Install ArgoCD CLI
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
rm argocd-linux-amd64

# Install Flux CLI
curl -s https://fluxcd.io/install.sh | sudo bash

# Install Kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
sudo mv kustomize /usr/local/bin/

# Install Tilt for local development
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

# Install k9s (Kubernetes TUI)
wget https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz
tar -xzf k9s_Linux_amd64.tar.gz
sudo mv k9s /usr/local/bin/
rm k9s_Linux_amd64.tar.gz

# Install yq (YAML processor)
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# Install kubectx and kubens
sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens

# Install Helm plugins
helm plugin install https://github.com/databus23/helm-diff || true
helm plugin install https://github.com/jkroepke/helm-secrets || true

# Install Python dependencies
echo "🐍 Installing Python dependencies..."
pip install --user --no-cache-dir \
  ansible==13.2.0 \
  ansible-lint==26.1.0 \
  black==25.12.0 \
  ruff==0.14.11 \
  pytest==8.4.2 \
  pytest-cov==7.0.0 \
  pre-commit==4.5.1 \
  yamllint==1.37.1

# Install pre-commit hooks
echo "🪝 Installing pre-commit hooks..."
pre-commit install || echo "No .pre-commit-config.yaml found, skipping"

# Create kind cluster for local development (optional)
if [ "${CREATE_KIND_CLUSTER:-false}" = "true" ]; then
  echo "🎯 Creating kind cluster..."
  kind create cluster --name mantl-dev --config=- <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
fi

# Add kubectl aliases
cat >> ~/.bashrc <<'EOF'

# Kubectl aliases
alias k=kubectl
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kgn='kubectl get nodes'
alias kdp='kubectl describe pod'
alias kds='kubectl describe service'
alias kdd='kubectl describe deployment'
alias kl='kubectl logs'
alias klf='kubectl logs -f'
alias kex='kubectl exec -it'
alias kctx='kubectx'
alias kns='kubens'

# Helm aliases
alias h=helm
alias hls='helm list'
alias hi='helm install'
alias hu='helm upgrade'

# Git aliases
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline -10'

# Add krew to PATH
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
EOF

echo "✅ Development environment setup complete!"
echo ""
echo "📚 Quick start:"
echo "  - Create local kind cluster: make install-dev"
echo "  - Access Kubernetes dashboard: kubectl proxy"
echo "  - Use k9s for interactive cluster management: k9s"
echo "  - Start hot-reload development: tilt up"
echo ""
echo "🔧 Installed tools:"
echo "  - kubectl, helm, kind, minikube"
echo "  - argocd, flux, kustomize"
echo "  - k9s, stern, kubectx, kubens"
echo "  - tilt, yq, jq"
echo "  - aws-cli, az-cli, gh-cli"
echo "  - terraform, python, node, go"
echo ""
echo "Type 'k get nodes' to verify kubectl is working!"
