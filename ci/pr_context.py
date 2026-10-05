"""Validate the exact PR head for bot-dispatched CI and Sonar analysis."""

import json
import os
import re
import subprocess
import sys


def context(pr, expected_head, repository, allow_fork):
    if pr.get("state") != "open":
        raise RuntimeError("CI requires an open pull request")
    head, base = pr["head"], pr["base"]
    if head["sha"] != expected_head:
        raise RuntimeError("PR head changed; dispatch CI on the current PR branch")
    if not allow_fork and head["repo"]["full_name"] != repository:
        raise RuntimeError("Manual PR CI requires a reviewed branch in this repository")
    return {
        "pull_request": str(pr["number"]),
        "head_branch": head["ref"],
        "base_branch": base["ref"],
        "base_sha": base["sha"],
    }


def main():
    number = os.environ.get("PR_NUMBER", "")
    if not number:
        return 0
    if not re.fullmatch(r"[1-9]\d*", number, flags=re.ASCII):
        print("::error::Invalid pull request number", file=sys.stderr)
        return 1
    repository = os.environ["GITHUB_REPOSITORY"]
    try:
        response = subprocess.run(
            ["gh", "api", f"repos/{repository}/pulls/{number}"],
            capture_output=True,
            text=True,
            check=True,
        )
        values = context(
            json.loads(response.stdout),
            os.environ["EXPECTED_HEAD"],
            repository,
            os.environ.get("ALLOW_FORK") == "true",
        )
    except subprocess.CalledProcessError:
        print("::error::Could not read pull request metadata", file=sys.stderr)
        return 1
    except (RuntimeError, KeyError, ValueError) as exc:
        print(f"::error::Invalid PR context: {exc}", file=sys.stderr)
        return 1
    for key, value in values.items():
        print(f"{key}={value}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
