# SPDX-License-Identifier: Apache-2.0
"""Exercise tool selection and release preflight without external services."""

import copy
import subprocess

import pytest

from ci import changed_components, release_pr_preflight, tool_changes


@pytest.mark.parametrize(
    "tools,expected",
    [
        ({"aqua:kyverno/kyverno"}, {"charts"}),
        ({"terraform"}, {"infrastructure", "developer_image"}),
        ({"go"}, {"runtime", "developer_image"}),
        ({"python"}, {"runtime", "developer_image", "ml_example", "python_examples"}),
        ({"goreleaser"}, {"runtime"}),
        ({"helm", "kubeconform"}, {"charts", "developer_image", "runtime"}),
        ({"golangci-lint"}, set()),
        (set(), set()),
    ],
)
def test_known_tools_select_consumers(tools, expected):
    result = changed_components.select(["mise.toml"], tools=tools)
    assert {name for name, selected in result.items() if selected} == expected


@pytest.mark.parametrize("tools", [None, {"unknown-new-tool"}])
def test_unknown_tools_keep_full_checks(tools):
    assert all(changed_components.select(["mise.toml"], tools=tools).values())
    assert len(changed_components.select_examples(["mise.toml"], tools=tools)) == 5


def test_tool_selection_keeps_content_checks():
    result = changed_components.select(
        ["mise.toml", "infra/terraform/blueprints/aws-eks/main.tf"], tools={"aqua:kyverno/kyverno"}
    )
    assert result["charts"]
    assert result["runtime"]
    assert result["infrastructure"]
    assert not result["frontend"]
    assert not result["python_examples"]


def test_full_examples_and_helper_changes_remain_complete():
    assert len(changed_components.select_examples(["mise.toml"], full=True, tools={"go"})) == 5
    assert all(changed_components.select(["ci/tool_changes.py"]).values())


def test_tool_diff_includes_addition_change_and_removal():
    assert tool_changes.changed_tools(
        '[tools]\ngo="1"\npython="2"', '[tools]\ngo="3"\nhelm="4"'
    ) == {"go", "python", "helm"}
    assert tool_changes.changed_tools('[tools]\ngo="1"', '# comment\n[tools]\ngo="1"') == set()


@pytest.mark.parametrize(
    "after",
    [
        '[env]\nMODE="new"\n[tools]\ngo="1"',
        "[tools",
        '[tools]\ngo="1"\n[settings]\nexperimental=true',
    ],
)
def test_broader_or_invalid_config_keeps_full_checks(after):
    assert tool_changes.changed_tools('[tools]\ngo="1"', after) is None


def test_git_reader_uses_exact_revisions(monkeypatch):
    responses = iter(['[tools]\ngo="1"', '[tools]\ngo="2"'])
    calls = []

    def run(args, **kwargs):
        calls.append((args, kwargs))
        return subprocess.CompletedProcess(args, 0, stdout=next(responses))

    monkeypatch.setattr(tool_changes.subprocess, "run", run)
    assert tool_changes.from_git("a" * 40, "b" * 40) == {"go"}
    assert calls[0][0] == ["git", "show", "a" * 40 + ":mise.toml"]
    assert all(call[1]["timeout"] == 20 for call in calls)
    assert tool_changes.from_git("main", "b" * 40) is None


@pytest.fixture
def release_pr():
    return {
        "number": 339,
        "state": "open",
        "base": {"ref": "main"},
        "head": {
            "sha": "a" * 40,
            "ref": "release-please--branches--main",
            "repo": {"full_name": "somethingwithproof/mantl"},
        },
    }


def run_for(pr, **changes):
    return {
        "event": "pull_request",
        "head_sha": pr["head"]["sha"],
        "head_branch": pr["head"]["ref"],
        "pull_requests": [{"number": pr["number"]}],
        **changes,
    }


@pytest.mark.parametrize(
    "status", ["queued", "in_progress", "completed", "action_required", "cancelled", "failure"]
)
def test_matching_pr_run_prevents_dispatch(release_pr, status):
    calls = []

    def read(endpoint):
        return (
            release_pr
            if "/pulls/" in endpoint
            else {"workflow_runs": [run_for(release_pr, status=status)]}
        )

    assert (
        release_pr_preflight.preflight(
            "somethingwithproof/mantl", 339, read, calls.append, lambda _: None
        )
        == "existing-pr-run"
    )
    assert calls == []


@pytest.mark.parametrize(
    "changes",
    [
        {"event": "workflow_dispatch"},
        {"head_sha": "b" * 40},
        {"head_branch": "unrelated"},
        {"pull_requests": [{"number": 123}]},
    ],
)
def test_unrelated_runs_do_not_suppress_preflight(release_pr, changes):
    assert not release_pr_preflight.matching_run([run_for(release_pr, **changes)], release_pr)


def test_missing_pr_run_dispatches_exact_branch_once(release_pr):
    calls, waits, reads = [], [], []

    def read(endpoint):
        reads.append(endpoint)
        return release_pr if "/pulls/" in endpoint else {"workflow_runs": []}

    assert (
        release_pr_preflight.preflight(
            "somethingwithproof/mantl", 339, read, calls.append, waits.append
        )
        == "advisory-preflight-dispatched"
    )
    assert len(calls) == 1
    assert waits == [5, 5]
    assert calls[0][-4:] == [
        "--ref",
        "release-please--branches--main",
        "--field",
        "pull-request=339",
    ]
    assert "head_sha=" + "a" * 40 in reads[1]


def test_run_appearing_during_grace_is_reused(release_pr):
    calls = []
    responses = iter([[], [run_for(release_pr)]])

    def read(endpoint):
        return release_pr if "/pulls/" in endpoint else {"workflow_runs": next(responses)}

    assert (
        release_pr_preflight.preflight(
            "somethingwithproof/mantl", 339, read, calls.append, lambda _: None
        )
        == "existing-pr-run"
    )
    assert calls == []


@pytest.mark.parametrize("state,sha", [("closed", "a" * 40), ("open", "b" * 40)])
def test_closed_or_changed_pr_is_not_dispatched(release_pr, state, sha):
    new = copy.deepcopy(release_pr)
    new["state"], new["head"]["sha"] = state, sha
    pulls = iter([release_pr, new])
    calls = []

    def read(endpoint):
        return next(pulls) if "/pulls/" in endpoint else {"workflow_runs": []}

    assert (
        release_pr_preflight.preflight(
            "somethingwithproof/mantl", 339, read, calls.append, lambda _: None
        )
        == "superseded"
    )
    assert calls == []


def test_fork_is_rejected_before_dispatch(release_pr):
    release_pr["head"]["repo"]["full_name"] = "someone/fork"
    with pytest.raises(ValueError, match="same-repository"):
        release_pr_preflight.preflight("somethingwithproof/mantl", 339, lambda _: release_pr)


@pytest.mark.parametrize(
    "repository,number",
    [
        ("bad/repo/extra", 339),
        ("somethingwithproof/mantl", "--help"),
        ("somethingwithproof/mantl", 0),
        ("somethingwithproof/mantl", "3٣9"),
    ],
)
def test_invalid_identity_never_calls_github(repository, number):
    def read(_):
        pytest.fail("invalid identity reached GitHub")

    with pytest.raises(ValueError):
        release_pr_preflight.preflight(repository, number, read)


def test_already_closed_pr_is_skipped(release_pr):
    release_pr["state"] = "closed"
    assert (
        release_pr_preflight.preflight("somethingwithproof/mantl", 339, lambda _: release_pr)
        == "closed"
    )


@pytest.mark.parametrize("field,value", [("number", 123), ("sha", "not-a-commit")])
def test_unexpected_api_identity_is_rejected(release_pr, field, value):
    if field == "number":
        release_pr[field] = value
    else:
        release_pr["head"][field] = value
    with pytest.raises(ValueError, match="identity"):
        release_pr_preflight.preflight("somethingwithproof/mantl", 339, lambda _: release_pr)


def test_api_read_is_bounded_and_dispatch_uses_exact_args(release_pr, monkeypatch):
    calls = []

    def run(args, **kwargs):
        calls.append((args, kwargs))
        return subprocess.CompletedProcess(args, 0, stdout='{"workflow_runs":[]}')

    monkeypatch.setattr(release_pr_preflight.subprocess, "run", run)
    assert release_pr_preflight.read_json("repos/somethingwithproof/mantl/actions/runs") == {
        "workflow_runs": []
    }
    assert calls[0][0][:2] == ["gh", "api"]
    assert calls[0][1]["timeout"] == 30

    def read(endpoint):
        return release_pr if "/pulls/" in endpoint else {"workflow_runs": []}

    release_pr_preflight.preflight("somethingwithproof/mantl", 339, read, sleep=lambda _: None)
    assert calls[-1][0][-4:] == [
        "--ref",
        "release-please--branches--main",
        "--field",
        "pull-request=339",
    ]
    assert calls[-1][1] == {"check": True, "timeout": 30}


def test_main_discloses_advisory_fallback(monkeypatch, capsys):
    monkeypatch.setenv("GITHUB_REPOSITORY", "somethingwithproof/mantl")
    monkeypatch.setenv("PR_NUMBER", "339")
    monkeypatch.setattr(
        release_pr_preflight, "preflight", lambda repo, pr: "advisory-preflight-dispatched"
    )
    release_pr_preflight.main()
    assert "required PR workflows still need approval" in capsys.readouterr().out


def test_missing_git_configuration_keeps_full_checks(monkeypatch):
    def run(*args, **kwargs):
        raise subprocess.TimeoutExpired(args[0], 20)

    monkeypatch.setattr(tool_changes.subprocess, "run", run)
    assert tool_changes.from_git("a" * 40, "b" * 40) is None
