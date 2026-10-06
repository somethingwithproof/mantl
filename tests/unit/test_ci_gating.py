"""Exercise CI selection and aggregate failures without cloud access."""

import importlib.util
import io
import os
import subprocess
from pathlib import Path

import pytest
import yaml

COMPONENTS = ["infrastructure", "charts", "runtime", "frontend", "developer_image"]


@pytest.fixture
def workflow(project_root: Path) -> dict:
    return yaml.safe_load((project_root / ".github/workflows/ci.yml").read_text())


@pytest.mark.parametrize(
    "required", ["CHANGES", "GO", "PYTHON", "PYTHON_EXAMPLES", "SONAR", "REPOSITORY"]
)
@pytest.mark.parametrize("result", ["failure", "cancelled", "skipped"])
def test_required_prerequisite_cannot_be_skipped(workflow, required, result):
    assert run_gate(workflow, {required + "_RESULT": result}).returncode != 0


@pytest.mark.parametrize("component", COMPONENTS)
@pytest.mark.parametrize(
    ("selected", "result", "passes"),
    [
        ("true", "success", True),
        ("true", "failure", False),
        ("true", "cancelled", False),
        ("true", "skipped", False),
        ("false", "skipped", True),
        ("false", "failure", False),
        ("", "skipped", False),
    ],
)
def test_selected_checks_must_finish(workflow, component, selected, result, passes):
    response = run_gate(
        workflow,
        {component.upper() + "_SELECTED": selected, component.upper() + "_RESULT": result},
    )
    assert (response.returncode == 0) is passes


def run_gate(workflow, changes):
    env = {
        **os.environ,
        **{
            key + "_RESULT": "success"
            for key in ["CHANGES", "GO", "PYTHON", "PYTHON_EXAMPLES", "SONAR", "REPOSITORY"]
        },
        **{key.upper() + "_SELECTED": "false" for key in COMPONENTS},
        **{key.upper() + "_RESULT": "skipped" for key in COMPONENTS},
        **changes,
    }
    script = workflow["jobs"]["ci"]["steps"][0]["run"]
    return subprocess.run(
        ["bash", "-euo", "pipefail", "-c", script],
        env=env,
        capture_output=True,
        text=True,
        check=False,
    )


@pytest.fixture
def selector(project_root):
    spec = importlib.util.spec_from_file_location(
        "changed_components", project_root / "ci/changed_components.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


@pytest.mark.parametrize(
    ("path", "expected"),
    [
        ("docs/architecture-runtime.md", set()),
        ("pkg/fleet/server.go", {"runtime"}),
        ("infra/terraform/blueprints/aws-eks/main.tf", {"infrastructure", "runtime"}),
        ("charts/mantl-platform/values.yaml", {"charts"}),
        ("Dockerfile", {"developer_image"}),
        ("requirements.in", {"developer_image"}),
        ("requirements-test.in", {"developer_image"}),
        (".devcontainer/post-create.sh", {"developer_image"}),
        ("scripts/release_paths.py", {"runtime"}),
        ("ci/verify-cli-packages.sh", {"runtime"}),
        ("examples/ml-inference-service/src/requirements.txt", {"ml_example"}),
        ("examples/ml-inference-service/src/Dockerfile", {"ml_example"}),
        ("ci/smoke_ml_example.py", {"ml_example"}),
        (".github/workflows/ci.yml", {"ml_example"}),
        ("Dockerfile.compliance-operator", {"runtime"}),
        ("examples/ecommerce-microservices/src/frontend/package-lock.json", {"frontend"}),
        (".github/workflows/release.yml", {"runtime"}),
        (".github/workflows/terraform-validate.yml", {"infrastructure"}),
        ("ci/validate-terraform.sh", {"infrastructure"}),
        (".github/actions/setup/action.yml", set(COMPONENTS) | {"ml_example"}),
        ("mise.toml", set(COMPONENTS) | {"ml_example"}),
    ],
)
def test_component_selection(selector, path, expected):
    assert {name for name, selected in selector.select([path]).items() if selected} == expected


def test_nul_delimited_paths_preserve_spaces(selector, monkeypatch, capsys):
    monkeypatch.setattr(
        selector.sys,
        "stdin",
        io.TextIOWrapper(io.BytesIO(b"docs/name with spaces.md\0pkg/new.go\0")),
    )
    selector.main()
    output = capsys.readouterr().out.splitlines()
    assert "runtime=true" in output
    assert "infrastructure=false" in output
