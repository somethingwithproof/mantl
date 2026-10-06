# SPDX-License-Identifier: Apache-2.0
"""Verify the configured SonarCloud project before generating coverage."""

import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path


def verify(token, properties, repository_visibility="private"):
    if not token:
        raise RuntimeError("SONAR_TOKEN is missing")
    project = properties.get("sonar.projectKey")
    if not project or not properties.get("sonar.organization"):
        raise RuntimeError("SonarCloud project key and organization must be configured")
    url = "https://sonarcloud.io/api/components/show?" + urllib.parse.urlencode(
        {"component": project}
    )
    request = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            data = json.load(response)
    except urllib.error.HTTPError as exc:
        raise RuntimeError(
            f"SonarCloud project verification returned HTTP {exc.code}; "
            "check token permissions and sonar.projectKey"
        ) from None
    except urllib.error.URLError:
        raise RuntimeError("SonarCloud project verification could not reach the service") from None
    component = data.get("component", {})
    if component.get("key") != project or component.get("qualifier") != "TRK":
        raise RuntimeError("SonarCloud returned an unexpected project")
    organization = component.get("organization")
    if organization and organization != properties["sonar.organization"]:
        raise RuntimeError("SonarCloud project belongs to a different organization")
    if repository_visibility not in {"private", "public"}:
        raise RuntimeError("Repository visibility is unknown; analysis upload refused")
    visibility = component.get("visibility")
    if visibility not in {"private", "public"} or (
        repository_visibility == "private" and visibility != "private"
    ):
        raise RuntimeError("Private source cannot use a public SonarCloud project; upload refused")
    return project


def main():
    properties = {}
    for line in Path("sonar-project.properties").read_text().splitlines():
        if line.strip() and not line.lstrip().startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            properties[key.strip()] = value.strip()
    try:
        project = verify(
            os.environ.get("SONAR_TOKEN"),
            properties,
            os.environ.get("MANTL_REPOSITORY_VISIBILITY") or "private",
        )
    except (RuntimeError, ValueError) as exc:
        print(f"::error::{exc}", file=sys.stderr)
        return 1
    print(f"Verified SonarCloud project: {project}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
