"""Select optional CI checks from NUL-delimited changed paths."""

import json
import os
import sys

EXAMPLES = (
    {"name": "publisher", "source": "examples/event-driven-arch/publisher"},
    {"name": "subscriber", "source": "examples/event-driven-arch/subscriber"},
    {"name": "products", "source": "examples/ecommerce-microservices/src/products-api"},
    {"name": "batch", "source": "examples/batch-processing-job/src"},
    {"name": "ml", "source": "examples/ml-inference-service/src"},
)


def executable_paths(paths):
    """Ignore reference prose, while retaining delivered framework/template content."""
    return [
        p
        for p in paths
        if p
        and (
            not p.endswith((".md", ".rst"))
            or p.startswith(("templates/", "compliance/frameworks/"))
        )
    ]


def select_examples(paths, full=False):
    """Select each affected example, or the complete baseline on main/manual runs."""
    paths = executable_paths(paths)
    shared = full or any(
        p.startswith((".github/actions/", "tests/examples/"))
        or p
        in {
            "mise.toml",
            "requirements.in",
            "requirements.txt",
            "requirements-test.in",
            "requirements-test.txt",
            "tests/conftest.py",
            "ci/changed_components.py",
            "ci/lock_python_dependencies.py",
            ".github/workflows/ci.yml",
            ".github/workflows/sonarcloud.yml",
        }
        for p in paths
    )
    return [
        example
        for example in EXAMPLES
        if shared
        or any(
            p.startswith(example["source"] + "/")
            or (example["name"] == "ml" and p == "ci/smoke_ml_example.py")
            for p in paths
        )
    ]


def select(paths, full_examples=False):
    paths = executable_paths(paths)
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
            "scripts/release_paths.py",
            "ci/verify-cli-packages.sh",
            "ci/verify_release_candidate.py",
        ),
        "frontend": ("examples/ecommerce-microservices/src/frontend/",),
        "ml_example": ("examples/ml-inference-service/src/", "ci/smoke_ml_example.py"),
        "developer_image": (
            "Dockerfile",
            "requirements.txt",
            "requirements-test.txt",
            "requirements.in",
            "requirements-test.in",
            ".devcontainer/",
        ),
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
    result["python_examples"] = bool(select_examples(paths, full_examples))
    return result


def main():
    paths = sys.stdin.buffer.read().decode().split("\0")
    full = os.environ.get("FULL_EXAMPLES", "false") == "true"
    selected = select(paths, full)
    for key, value in selected.items():
        print(f"{key}={str(value).lower()}")
    print("example_matrix=" + json.dumps(select_examples(paths, full), separators=(",", ":")))


if __name__ == "__main__":
    main()
