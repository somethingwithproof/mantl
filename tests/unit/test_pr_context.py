"""Dispatched CI must analyze the current, approved repository branch."""

import importlib.util
import json
import subprocess

import pytest


@pytest.fixture
def module(project_root):
    spec = importlib.util.spec_from_file_location("pr_context", project_root / "ci/pr_context.py")
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


@pytest.fixture
def pr():
    return {
        "number": 42,
        "state": "open",
        "head": {"sha": "head", "ref": "release-branch", "repo": {"full_name": "owner/repo"}},
        "base": {"sha": "base", "ref": "main"},
    }


def test_current_head_resolves_sonar_context(module, pr):
    assert module.context(pr, "head", "owner/repo", False) == {
        "pull_request": "42",
        "head_branch": "release-branch",
        "base_branch": "main",
        "base_sha": "base",
    }


def test_stale_head_cannot_pass_checks_for_updated_pr(module, pr):
    with pytest.raises(RuntimeError, match="PR head changed"):
        module.context(pr, "previous-head", "owner/repo", False)


def test_closed_pr_is_rejected(module, pr):
    pr["state"] = "closed"
    with pytest.raises(RuntimeError, match="open pull request"):
        module.context(pr, "head", "owner/repo", False)


def test_manual_dispatch_cannot_use_unreviewed_fork(module, pr):
    pr["head"]["repo"]["full_name"] = "fork/repo"
    with pytest.raises(RuntimeError, match="reviewed branch"):
        module.context(pr, "head", "owner/repo", False)
    assert module.context(pr, "head", "owner/repo", True)["pull_request"] == "42"


def test_non_pr_ci_does_not_contact_github(module, monkeypatch):
    monkeypatch.delenv("PR_NUMBER", raising=False)
    monkeypatch.setattr(module.subprocess, "run", lambda *_a, **_k: pytest.fail("request"))
    assert module.main() == 0


@pytest.mark.parametrize("number", ["../other", "4٢", "0", "42\n"])
def test_invalid_number_does_not_contact_github(module, monkeypatch, number):
    monkeypatch.setenv("PR_NUMBER", number)
    monkeypatch.setattr(module.subprocess, "run", lambda *_a, **_k: pytest.fail("request"))
    assert module.main() == 1


def test_cli_emits_only_validated_metadata(module, pr, monkeypatch, capsys):
    monkeypatch.setenv("PR_NUMBER", "42")
    monkeypatch.setenv("GITHUB_REPOSITORY", "owner/repo")
    monkeypatch.setenv("EXPECTED_HEAD", "head")
    monkeypatch.setattr(
        module.subprocess,
        "run",
        lambda *_a, **_k: subprocess.CompletedProcess([], 0, json.dumps(pr)),
    )
    assert module.main() == 0
    assert "pull_request=42" in capsys.readouterr().out
    monkeypatch.setenv("EXPECTED_HEAD", "stale")
    assert module.main() == 1
    assert "PR head changed" in capsys.readouterr().err


def test_github_failure_is_sanitized(module, monkeypatch, capsys):
    monkeypatch.setenv("PR_NUMBER", "42")
    monkeypatch.setenv("GITHUB_REPOSITORY", "owner/repo")

    def fail(*_a, **_k):
        raise subprocess.CalledProcessError(1, [], stderr="sensitive response")

    monkeypatch.setattr(module.subprocess, "run", fail)
    assert module.main() == 1
    output = capsys.readouterr()
    assert "sensitive" not in output.out + output.err
