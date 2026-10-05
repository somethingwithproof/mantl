"""Select optional CI checks from NUL-delimited changed paths."""

import sys


def select(paths):
    shared = any(p.startswith(".github/actions/") or p == "mise.toml" for p in paths)
    prefixes = {
        "infrastructure": ("infra/terraform/", ".tflint.hcl", "ci/validate-terraform.sh"),
        "charts": (
            "charts/",
            "deploy/",
            "clusters/",
            "apps/examples/",
            "apps/templates/",
            "policies/",
            "compliance/frameworks/",
        ),
        "runtime": (
            "apis/",
            "cmd/",
            "controllers/",
            "pkg/",
            "integration/",
            "go.",
            "compliance/",
            "deploy/",
            "infra/terraform/",
            "templates/backstage/",
            "Dockerfile.compliance-operator",
            ".goreleaser.yaml",
            "scripts/package_release.py",
            "scripts/publish_control_bundle.py",
        ),
        "frontend": ("examples/ecommerce-microservices/src/frontend/",),
        "ml_example": ("examples/ml-inference-service/src/", "ci/smoke_ml_example.py"),
        "developer_image": ("Dockerfile", "requirements.txt", "requirements-test.txt"),
    }
    result = {}
    for component, roots in prefixes.items():
        workflow = {
            "infrastructure": "terraform-validate",
            "charts": "helm-validate",
            "runtime": "compliance-runtime",
            "frontend": "example-frontend",
            "ml_example": "ci",
            "developer_image": "developer-image",
        }[component]
        result[component] = shared or any(
            any(
                p.startswith(root) if root.endswith("/") or root == "go." else p == root
                for root in roots
            )
            or p == f".github/workflows/{workflow}.yml"
            or p == "ci/changed_components.py"
            or (component == "runtime" and p.startswith(".github/workflows/release"))
            for p in paths
        )
    return result


def main():
    paths = sys.stdin.buffer.read().decode().split("\0")
    selected = select(paths)
    for key, value in selected.items():
        print(f"{key}={str(value).lower()}")


if __name__ == "__main__":
    main()
