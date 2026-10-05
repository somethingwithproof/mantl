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
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", source)
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
    first, second = source / "dist/one", source / "dist/two"
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
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", tmp_path)
    archive = tmp_path / "dist/controls.tar.gz"
    archive.parent.mkdir()
    archive.write_bytes(b"fixture")
    calls = []

    def run(args, **_):
        calls.append(args)
        return subprocess.CompletedProcess(args, 1, "", "unauthorized")

    monkeypatch.setattr(module.subprocess, "run", run)
    with pytest.raises(RuntimeError, match="publication refused"):
        module.publish("ghcr.io/somethingwithproof/mantl-controls", "1.0.0", archive)
    assert len(calls) == 1
    assert calls[0][1] == "resolve"


def test_release_tooling_rejects_paths_before_external_commands(tmp_path, monkeypatch):
    module = load_script("package_release")
    root = tmp_path / "checkout"
    root.mkdir()
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", root)

    def unexpected_command(*_args, **_kwargs):
        pytest.fail("invalid paths must be rejected before external commands")

    monkeypatch.setattr(module.subprocess, "check_output", unexpected_command)
    image = "ghcr.io/somethingwithproof/mantl-operator@sha256:" + "0" * 64
    with pytest.raises(ValueError, match="source must"):
        module.assemble(tmp_path, root / "dist/release", "v1.0.0", image)
    with pytest.raises(ValueError, match="inside this checkout"):
        module.assemble(root, tmp_path / "outside", "v1.0.0", image)
    assert not (tmp_path / "outside").exists()


def test_release_paths_reject_noncanonical_and_symlinked_outputs(tmp_path, monkeypatch):
    module = load_script("package_release")
    paths = module.release_paths
    monkeypatch.setattr(paths, "WORKSPACE_ROOT", tmp_path)
    dist = tmp_path / "dist"
    dist.mkdir()
    assert paths.artifact_path(Path("dist/release")) == dist / "release"
    assert paths.source_path(tmp_path) == tmp_path
    with pytest.raises(ValueError, match="canonical"):
        paths.artifact_path(Path("dist/../outside"))
    link = dist / "link"
    link.symlink_to(tmp_path / "outside")
    with pytest.raises(ValueError, match="symlinks"):
        paths.artifact_path(link)
    with pytest.raises(ValueError, match="inside this checkout"):
        paths.artifact_path(link / "file")


def test_publication_rejects_external_or_missing_archive(tmp_path, monkeypatch):
    module = load_script("publish_control_bundle")
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", tmp_path)
    with pytest.raises(ValueError, match="inside this checkout"):
        module.publish("ghcr.io/somethingwithproof/mantl-controls", "1.0.0", tmp_path / "file")
    with pytest.raises(ValueError, match="regular file"):
        module.publish(
            "ghcr.io/somethingwithproof/mantl-controls", "1.0.0", tmp_path / "dist/missing"
        )


@pytest.mark.parametrize("tag", ["v١.0.0", "v1.0.0/extra", "1.0.0"])
def test_release_tags_remain_ascii_semver(tmp_path, monkeypatch, tag):
    module = load_script("package_release")
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", tmp_path)
    with pytest.raises(ValueError, match="SemVer"):
        module.assemble(
            tmp_path,
            tmp_path / "dist/release",
            tag,
            "ghcr.io/somethingwithproof/mantl-operator@sha256:" + "0" * 64,
        )
