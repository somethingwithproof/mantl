# SPDX-License-Identifier: Apache-2.0
"""Select optional CI checks from NUL-delimited changed paths."""

import json
import os
import sys

from ci import tool_changes

# Additional checks for each known tool. Core tests/hooks/Sonar always run.
TOOL_COMPONENTS = {
    "go": {"runtime", "developer_image"},
    "python": {"runtime", "developer_image", "ml_example", "python_examples"},
    "terraform": {"infrastructure", "developer_image"},
    "tflint": {"infrastructure"},
    "terraform-docs": {"infrastructure"},
    "helm": {"charts", "developer_image"},
    "kustomize": {"charts", "developer_image"},
    "kubectl": {"runtime", "developer_image"},
    "kubeconform": {"charts", "runtime"},
    "goreleaser": {"runtime"},
    "cosign": {"runtime"},
    "oras": {"runtime"},
    "syft": {"runtime"},
    "uv": {"developer_image"},
    "trivy": {"infrastructure"},
    "kyverno": {"charts"},
    "aqua:kyverno/kyverno": {"charts"},
    "golangci-lint": set(),
    "pre-commit": set(),
    "actionlint": set(),
    "shellcheck": set(),
}


def tool_components(paths, tools):
    if "mise.toml" not in paths:
        return set()
    if tools is None or any(name not in TOOL_COMPONENTS for name in tools):
        return None
    return set().union(*(TOOL_COMPONENTS[name] for name in tools))


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


def select_examples(paths, full=False, tools=None):
    """Select each affected example, or the complete baseline on main/manual runs."""
    paths = executable_paths(paths)
    affected = tool_components(paths, tools)
    shared = (
        full
        or affected is None
        or "python_examples" in affected
        or any(
            p.startswith((".github/actions/", "tests/examples/"))
            or p
            in {
                "requirements.in",
                "requirements.txt",
                "requirements-test.in",
                "requirements-test.txt",
                "tests/conftest.py",
                "ci/changed_components.py",
                "ci/tool_changes.py",
                "ci/lock_python_dependencies.py",
                ".github/workflows/ci.yml",
                ".github/workflows/sonarcloud.yml",
            }
            for p in paths
        )
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


def select(paths, full_examples=False, tools=None):
    paths = executable_paths(paths)
    affected = tool_components(paths, tools)
    shared = affected is None or any(
        p.startswith(".github/actions/") or p == "ci/tool_changes.py" for p in paths
    )
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
        result[component] = (
            shared
            or component in (affected or set())
            or any(
                any(
                    p.startswith(root) if root.endswith("/") or root == "go." else p == root
                    for root in roots
                )
                or p == f".github/workflows/{workflow}.yml"
                or p == "ci/changed_components.py"
                or (component == "runtime" and p.startswith(".github/workflows/release"))
                for p in paths
            )
        )
    result["python_examples"] = bool(select_examples(paths, full_examples, tools))
    return result


def main():
    paths = sys.stdin.buffer.read().decode().split("\0")
    full = os.environ.get("FULL_EXAMPLES", "false") == "true"
    tools = (
        tool_changes.from_git(os.environ.get("BASE"), os.environ.get("HEAD"))
        if "mise.toml" in paths
        else set()
    )
    selected = select(paths, full, tools)
    for key, value in selected.items():
        print(f"{key}={str(value).lower()}")
    print(
        "example_matrix=" + json.dumps(select_examples(paths, full, tools), separators=(",", ":"))
    )


if __name__ == "__main__":
    main()
