# SPDX-License-Identifier: Apache-2.0
"""Avoid parallel advisory CI when an eligible release PR run already exists."""

import json
import os
import re
import subprocess
import time


def matching_run(runs, pr):
    return any(
        run.get("event") == "pull_request"
        and run.get("head_sha") == pr["head"]["sha"]
        and run.get("head_branch") == pr["head"]["ref"]
        and any(item.get("number") == pr["number"] for item in run.get("pull_requests", []))
        for run in runs
    )


def read_json(endpoint):
    result = subprocess.run(
        ["gh", "api", endpoint], capture_output=True, text=True, check=True, timeout=30
    )
    return json.loads(result.stdout)


def preflight(repository, number, read=read_json, dispatch=None, sleep=time.sleep):
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repository):
        raise ValueError("invalid repository")
    if not re.fullmatch(r"[1-9]\d*", str(number), flags=re.ASCII):
        raise ValueError("invalid release PR number")
    endpoint = f"repos/{repository}/pulls/{number}"
    pr = read(endpoint)
    if pr["state"] != "open":
        return "closed"
    if pr["head"]["repo"]["full_name"] != repository or pr["base"]["ref"] != "main":
        raise ValueError("preflight requires a same-repository PR targeting main")
    head = pr["head"]["sha"]
    if not re.fullmatch(r"[a-f0-9]{40}", head) or pr["number"] != int(number):
        raise ValueError("invalid release PR identity")
    query = (
        f"repos/{repository}/actions/workflows/ci.yml/runs"
        f"?event=pull_request&head_sha={head}&per_page=100"
    )
    for attempt in range(3):
        if matching_run(read(query)["workflow_runs"], pr):
            return "existing-pr-run"
        if attempt < 2:
            sleep(5)
    current = read(endpoint)
    if current["state"] != "open" or current["head"]["sha"] != head:
        return "superseded"
    dispatch_advisory(repository, pr["head"]["ref"], number, dispatch)
    return "advisory-preflight-dispatched"


def dispatch_advisory(repository, branch, number, dispatch):
    args = [
        "gh",
        "workflow",
        "run",
        "ci.yml",
        "--repo",
        repository,
        "--ref",
        branch,
        "--field",
        f"pull-request={number}",
    ]
    if dispatch is None:
        subprocess.run(args, check=True, timeout=30)
    else:
        dispatch(args)


def main():
    result = preflight(os.environ["GITHUB_REPOSITORY"], os.environ["PR_NUMBER"])
    print(f"Release PR CI: {result}")
    if result == "advisory-preflight-dispatched":
        print(
            "::notice::Manual preflight is advisory; "
            "required PR workflows still need approval and successful checks."
        )


if __name__ == "__main__":
    main()
