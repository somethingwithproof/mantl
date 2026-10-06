# SPDX-License-Identifier: Apache-2.0
"""Exercise opt-in STIG workload policies through Kyverno's offline engine."""

import copy
import shutil
import subprocess
from pathlib import Path

import pytest
import yaml

POLICIES = Path(__file__).resolve().parents[2] / "policies/kyverno/stig-kubernetes"


@pytest.mark.parametrize("container_kind", ["containers", "initContainers", "ephemeralContainers"])
@pytest.mark.parametrize(
    "policy,field,allowed",
    [
        (
            "stig-disallow-secret-env",
            {
                "env": [
                    {
                        "name": "PASSWORD",
                        "valueFrom": {"secretKeyRef": {"name": "db", "key": "password"}},
                    }
                ]
            },
            False,
        ),
        ("stig-disallow-secret-env", {"envFrom": [{"secretRef": {"name": "db"}}]}, False),
        ("stig-disallow-secret-env", {"env": [{"name": "MODE", "value": "normal"}]}, True),
        ("stig-disallow-secret-env", {"envFrom": [{"configMapRef": {"name": "settings"}}]}, True),
        (
            "stig-restrict-privileged-host-ports",
            {"ports": [{"containerPort": 8080, "hostPort": 1023}]},
            False,
        ),
        (
            "stig-restrict-privileged-host-ports",
            {"ports": [{"containerPort": 8080, "hostPort": 1}]},
            False,
        ),
        (
            "stig-restrict-privileged-host-ports",
            {"ports": [{"containerPort": 80, "hostPort": 1024}]},
            True,
        ),
        (
            "stig-restrict-privileged-host-ports",
            {"ports": [{"containerPort": 80, "hostPort": 0}]},
            True,
        ),
        ("stig-restrict-privileged-host-ports", {"ports": [{"containerPort": 80}]}, True),
    ],
)
def test_stig_workload_policy(tmp_path, container_kind, policy, field, allowed):
    output = apply_policy(tmp_path, container_kind, policy, field, "team-a")
    assert ("pass: 1" if allowed else "warn: 1") in output, output


@pytest.mark.parametrize("container_kind", ["containers", "initContainers", "ephemeralContainers"])
@pytest.mark.parametrize(
    "namespace", ["kube-system", "kube-node-lease", "kube-public", "default", "team-a"]
)
def test_host_port_policy_stig_namespace_scope(tmp_path, container_kind, namespace):
    output = apply_policy(
        tmp_path,
        container_kind,
        "stig-restrict-privileged-host-ports",
        {"ports": [{"containerPort": 80, "hostPort": 80}]},
        namespace,
    )
    expected = "pass: 0, fail: 0, warn: 0, error: 0" if namespace.startswith("kube-") else "warn: 1"
    assert expected in output, output


@pytest.mark.parametrize("container_kind", ["containers", "initContainers", "ephemeralContainers"])
@pytest.mark.parametrize("namespace", ["kube-system", "kube-node-lease", "kube-public"])
def test_secret_env_policy_still_audits_system_namespaces(tmp_path, container_kind, namespace):
    output = apply_policy(
        tmp_path,
        container_kind,
        "stig-disallow-secret-env",
        {"envFrom": [{"secretRef": {"name": "db"}}]},
        namespace,
    )
    assert "warn: 1" in output, output


def apply_policy(tmp_path, container_kind, policy, field, namespace):
    executable = shutil.which("kyverno")
    if executable is None:
        pytest.skip("Kyverno CLI required for offline policy execution")
    pod = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "app", "namespace": namespace},
        "spec": {"containers": [{"name": "app", "image": "example/app:1"}]},
    }
    container = {"name": "subject", "image": "example/app:1", **copy.deepcopy(field)}
    pod["spec"][container_kind] = [container]
    resource = tmp_path / "pod.yaml"
    resource.write_text(yaml.safe_dump(pod))
    result = subprocess.run(
        [
            executable,
            "apply",
            str(POLICIES / f"{policy}.yaml"),
            "--resource",
            str(resource),
            "--audit-warn",
        ],
        capture_output=True,
        text=True,
        timeout=30,
        check=False,
    )
    output = result.stdout + result.stderr
    assert result.returncode in (0, 2), output
    return output
