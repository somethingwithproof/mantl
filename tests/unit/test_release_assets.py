"""Release assets must exclude workstation state and preserve immutable identity."""

import importlib.util
import subprocess
import tarfile
from pathlib import Path

import pytest


def load_script(name):
    spec = importlib.util.spec_from_file_location(
        name, Path(__file__).parents[2] / "scripts" / f"{name}.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def test_platform_bundle_excludes_local_state_and_is_reproducible(tmp_path, monkeypatch):
    module = load_script("package_release")
    source = tmp_path / "source"
    tracked = [
        "infra/terraform/main.tf",
        "infra/terraform/terraform.tfvars.json",
        "infra/terraform/local.tfstate",
        "infra/terraform/private.key",
        "infra/terraform/.terraform/cache",
        "deploy/operator/base/deployment.yaml",
    ]
    for name in tracked + ["infra/terraform/untracked.env"]:
        file = source / name
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_text("fixture")
    (source / "infra/terraform/.terraform/provider-link").symlink_to(tmp_path / "outside")

    def output(args, **_):
        if args[:2] == ["git", "ls-files"]:
            return "\0".join(tracked).encode()
        if args[0] == "git":
            return "a" * 40
        return (
            b"image: mantl-compliance-operator:development\n"
            b"args: [--collector-image=mantl-compliance-operator:development]\n"
        )

    monkeypatch.setattr(module.subprocess, "check_output", output)
    image = "ghcr.io/somethingwithproof/mantl-operator@sha256:" + "0" * 64
    first, second = tmp_path / "one", tmp_path / "two"
    module.assemble(source, first, "v1.0.0", image)
    module.assemble(source, second, "v1.0.0", image)
    archive = first / "mantl-platform_v1.0.0.tar.gz"
    assert archive.read_bytes() == (second / archive.name).read_bytes()
    with tarfile.open(archive) as bundle:
        assert set(bundle.getnames()) == {
            "infra/terraform/main.tf",
            "deploy/operator/base/deployment.yaml",
        }
    rendered = (first / "operator-install.yaml").read_text()
    assert rendered.count(image) == 2
    assert "development" not in rendered


def test_content_publication_does_not_turn_auth_failure_into_push(tmp_path, monkeypatch):
    module = load_script("publish_control_bundle")
    calls = []

    def run(args, **_):
        calls.append(args)
        return subprocess.CompletedProcess(args, 1, "", "unauthorized")

    monkeypatch.setattr(module.subprocess, "run", run)
    with pytest.raises(RuntimeError, match="publication refused"):
        module.publish(
            "ghcr.io/somethingwithproof/mantl-controls", "1.0.0", tmp_path / "controls.tar.gz"
        )
    assert len(calls) == 1
    assert calls[0][1] == "resolve"
