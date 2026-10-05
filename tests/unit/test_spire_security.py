"""Keep node attestation privileges scoped and kubelet identity verified."""

import yaml


def test_spire_kubelet_trust_and_authorization(project_root):
    docs = list(
        yaml.safe_load_all((project_root / "infrastructure/cilium/mtls-config.yaml").read_text())
    )
    by_kind_name = {(doc["kind"], doc["metadata"]["name"]): doc for doc in docs if doc}
    config = by_kind_name[("ConfigMap", "spire-agent")]["data"]["agent.conf"]
    assert "skip_kubelet_verification = false" in config
    assert 'kubelet_ca_path = "/run/spire/kubelet-ca/ca.crt"' in config
    assert "kubelet_secure_port = 10250" in config
    agent = by_kind_name[("DaemonSet", "spire-agent")]["spec"]["template"]["spec"]
    volume = next(v for v in agent["volumes"] if v["name"] == "kubelet-ca")
    assert volume["configMap"] == {"name": "spire-kubelet-ca", "optional": False}
    mount = next(v for v in agent["containers"][0]["volumeMounts"] if v["name"] == "kubelet-ca")
    assert mount["readOnly"]
    assert mount["mountPath"] == "/run/spire/kubelet-ca"
    rules = by_kind_name[("ClusterRole", "spire-agent-cluster-role")]["rules"]
    for rule in rules:
        assert rule["verbs"] == ["get"]
        assert set(rule["resources"]) <= {"pods", "nodes", "nodes/pods"}
    assert any("nodes/pods" in rule["resources"] for rule in rules)
