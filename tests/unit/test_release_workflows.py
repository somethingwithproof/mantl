# SPDX-License-Identifier: Apache-2.0
"""Exercise release identity guards and metadata-only merge races offline."""

import os
import subprocess
from pathlib import Path

import pytest
import yaml

ROOT = Path(__file__).parents[2]


def workflow_script(filename, job, step):
    workflow = yaml.safe_load((ROOT / ".github/workflows" / filename).read_text())
    return next(item["run"] for item in workflow["jobs"][job]["steps"] if item.get("name") == step)


@pytest.mark.parametrize(
    ("ref", "sha_matches", "accepted"),
    [
        ("refs/heads/main", True, True),
        ("refs/tags/v1.0.0", True, True),
        ("refs/heads/main", False, False),
        ("refs/heads/feature", True, False),
        ("refs/tags/v2.0.0", True, False),
    ],
)
def test_release_identity_matches_provenance_source(tmp_path, ref, sha_matches, accepted):
    def git(*args):
        return subprocess.check_output(
            ["git", "-c", "core.hooksPath=/dev/null", "-c", "tag.gpgSign=false", *args],
            cwd=tmp_path,
            text=True,
        ).strip()

    git("init", "-b", "main")
    git(
        "-c",
        "user.name=Fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "-c",
        "commit.gpgsign=false",
        "commit",
        "--allow-empty",
        "-m",
        "fixture",
    )
    sha = git("rev-parse", "HEAD")
    git("tag", "v1.0.0")
    git("update-ref", "refs/remotes/origin/main", sha)
    output = tmp_path / "output"
    result = subprocess.run(
        [
            "bash",
            "-c",
            workflow_script("release.yml", "build", "Validate immutable release identity"),
        ],
        cwd=tmp_path,
        env={
            **os.environ,
            "RELEASE_TAG": "v1.0.0",
            "GITHUB_REF": ref,
            "GITHUB_SHA": sha if sha_matches else "0" * 40,
            "GITHUB_OUTPUT": str(output),
        },
        capture_output=True,
        text=True,
    )
    assert (result.returncode == 0) == accepted
    assert output.exists() == accepted


@pytest.mark.parametrize("scenario", ["race", "denied", "closed", "churn", "major"])
def test_dependabot_sweep_continues_after_first_pr_failure(tmp_path, scenario):
    gh = tmp_path / "gh"
    gh.write_text(
        """#!/bin/bash
set -eu
case "$2" in
  list) printf '1 old\n2 stable\n' ;;
  --paginate)
    if test "$SCENARIO" = major && [[ "$*" == *pulls/1/* ]]; then
      echo "update-type: version-update:semver-major"
    fi ;;
  merge)
    echo "$*" >> "$CALLS"
    if test "$3" = 2; then exit 0; fi
    if test "$SCENARIO" = race && test "${!#}" = new; then exit 0; fi
    exit 1 ;;
  view)
    case "$SCENARIO" in
      race) echo '{"state":"OPEN","isDraft":false,"headRefOid":"new"}' ;;
      denied) echo '{"state":"OPEN","isDraft":false,"headRefOid":"old"}' ;;
      closed) echo '{"state":"MERGED","isDraft":false,"headRefOid":"old"}' ;;
      churn) printf '{"state":"OPEN","isDraft":false,"headRefOid":"%s"}\n' "$(wc -l < "$CALLS")" ;;
      major) echo '{"state":"OPEN","isDraft":false,"headRefOid":"old"}' ;;
    esac ;;
esac
"""
    )
    gh.chmod(0o755)
    calls = tmp_path / "calls"
    script = workflow_script(
        "dependabot-auto-merge.yml", "merge", "Enable protected Dependabot auto-merge"
    )
    # Isolate the hosted runner's scratch file from other local test processes.
    script = script.replace("/tmp/dependabot-prs.tsv", str(tmp_path / "prs.tsv"))
    result = subprocess.run(
        ["bash", "-c", script],
        env={
            **os.environ,
            "PATH": f"{tmp_path}:{os.environ['PATH']}",
            "GITHUB_REPOSITORY": "owner/repo",
            "CALLS": str(calls),
            "SCENARIO": scenario,
        },
        capture_output=True,
        text=True,
    )
    commands = calls.read_text().splitlines()
    assert commands[-1].startswith("pr merge 2 ")
    assert all("--auto --squash --match-head-commit" in command for command in commands)
    assert len(commands) == {"race": 3, "denied": 2, "closed": 2, "churn": 4, "major": 1}[scenario]
    assert (result.returncode == 0) == (scenario in {"race", "closed", "major"})
