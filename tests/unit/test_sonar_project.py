"""Reject invalid SonarCloud access without disclosing credentials."""

import io
import urllib.error

import pytest
from test_release_assets import load_script

PROPERTIES = {"sonar.projectKey": "fixture_project", "sonar.organization": "fixture_org"}


def test_missing_token_never_contacts_service(monkeypatch):
    module = load_script("verify_sonar_project")
    monkeypatch.setattr(module.urllib.request, "urlopen", lambda *_a, **_k: pytest.fail("request"))
    with pytest.raises(RuntimeError, match="SONAR_TOKEN is missing"):
        module.verify(None, PROPERTIES)


@pytest.mark.parametrize("status", [401, 403, 404])
def test_authentication_or_project_failure_is_sanitized(monkeypatch, status):
    module = load_script("verify_sonar_project")

    def respond(request, **_):
        raise urllib.error.HTTPError(request.full_url, status, "", {}, io.BytesIO(b"sensitive"))

    monkeypatch.setattr(module.urllib.request, "urlopen", respond)
    with pytest.raises(RuntimeError, match=f"HTTP {status}") as error:
        module.verify("fixture-token", PROPERTIES)
    assert "sensitive" not in str(error.value)
    assert "fixture-token" not in str(error.value)


@pytest.mark.parametrize(
    "component",
    [
        {"key": "wrong_project", "qualifier": "TRK"},
        {"key": "fixture_project", "qualifier": "DIR"},
        {"key": "fixture_project", "qualifier": "TRK", "organization": "wrong_org"},
    ],
)
def test_wrong_project_or_organization_is_rejected(monkeypatch, component):
    module = load_script("verify_sonar_project")
    monkeypatch.setattr(module.urllib.request, "urlopen", lambda *_a, **_k: io.BytesIO(b"{}"))
    monkeypatch.setattr(module.json, "load", lambda _: {"component": component})
    with pytest.raises(RuntimeError):
        module.verify("fixture-token", PROPERTIES)


def test_matching_project_is_verified(monkeypatch):
    module = load_script("verify_sonar_project")
    monkeypatch.setattr(module.urllib.request, "urlopen", lambda *_a, **_k: io.BytesIO(b"{}"))
    monkeypatch.setattr(
        module.json,
        "load",
        lambda _: {
            "component": {"key": "fixture_project", "qualifier": "TRK", "visibility": "private"}
        },
    )
    assert module.verify("fixture-token", PROPERTIES) == "fixture_project"


@pytest.mark.parametrize("visibility", ["public", None])
def test_nonprivate_analysis_project_is_rejected(monkeypatch, visibility):
    module = load_script("verify_sonar_project")
    monkeypatch.setattr(module.urllib.request, "urlopen", lambda *_a, **_k: io.BytesIO(b"{}"))
    monkeypatch.setattr(
        module.json,
        "load",
        lambda _: {
            "component": {"key": "fixture_project", "qualifier": "TRK", "visibility": visibility}
        },
    )
    with pytest.raises(RuntimeError, match="upload refused"):
        module.verify("fixture-token", PROPERTIES)


def test_public_repository_can_use_public_analysis(monkeypatch):
    module = load_script("verify_sonar_project")
    monkeypatch.setattr(module.urllib.request, "urlopen", lambda *_a, **_k: io.BytesIO(b"{}"))
    monkeypatch.setattr(
        module.json,
        "load",
        lambda _: {
            "component": {"key": "fixture_project", "qualifier": "TRK", "visibility": "public"}
        },
    )
    assert module.verify("fixture-token", PROPERTIES, "public") == "fixture_project"
