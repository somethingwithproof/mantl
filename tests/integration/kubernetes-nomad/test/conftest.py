import json
import os
import time

import consul
import hvac
import pytest
import requests
from kubernetes import client, config

# Load Kubernetes configuration


def pytest_configure(config):
    """Set up test environment and connections to services."""
    control_node = os.environ.get('CONTROL_NODE', 'control')

    # Set up global variables
    global NOMAD_ADDR, CONSUL_ADDR, VAULT_ADDR, K8S_ADDR

    NOMAD_ADDR = f"http://{control_node}:4646"
    CONSUL_ADDR = f"http://{control_node}:8500"
    VAULT_ADDR = f"http://{control_node}:8200"
    K8S_ADDR = f"https://{control_node}:6443"

    # Load kubernetes config
    try:
        # First try loading from shared kubeconfig
        if os.path.exists('/etc/mantl/config/kubeconfig'):
            config.load_kube_config(config_file='/etc/mantl/config/kubeconfig')
        else:
            # Fall back to default location
            config.load_kube_config()
    except Exception as e:
        print(f"Error loading Kubernetes config: {e}")
        # For testing in Docker, try in-cluster config
        try:
            config.load_incluster_config()
        except Exception as e2:
            print(f"Error loading in-cluster config: {e2}")
            # Last resort: use environment variables
            os.environ['KUBERNETES_SERVICE_HOST'] = control_node
            os.environ['KUBERNETES_SERVICE_PORT'] = '6443'


@pytest.fixture
def nomad_client():
    """Fixture for Nomad HTTP API client."""
    class NomadClient:
        def __init__(self):
            self.base_url = NOMAD_ADDR
            self.headers = {"Content-Type": "application/json"}

        def get(self, endpoint):
            resp = requests.get(f"{self.base_url}/v1/{endpoint}", headers=self.headers)
            resp.raise_for_status()
            return resp.json()

        def post(self, endpoint, data):
            resp = requests.post(f"{self.base_url}/v1/{endpoint}",
                                 headers=self.headers,
                                 data=json.dumps(data))
            resp.raise_for_status()
            return resp.json()

    return NomadClient()


@pytest.fixture
def consul_client():
    """Fixture for Consul client."""
    control_node = os.environ.get('CONTROL_NODE', 'control')
    return consul.Consul(host=control_node, port=8500)


@pytest.fixture
def vault_client():
    """Fixture for Vault client."""
    control_node = os.environ.get('CONTROL_NODE', 'control')
    client = hvac.Client(url=f"http://{control_node}:8200")

    # Try to get token from environment or file
    token = os.environ.get('VAULT_TOKEN')
    if not token:
        try:
            with open('/etc/mantl/test/vault-token.txt', 'r') as f:
                token = f.read().strip()
        except BaseException:
            # Just use root token for testing
            with open('/var/lib/vault/init.txt', 'r') as f:
                for line in f:
                    if "Initial Root Token" in line:
                        token = line.split(": ")[1].strip()

    if token:
        client.token = token

    return client


@pytest.fixture
def k8s_api_client():
    """Fixture for Kubernetes API client."""
    config.load_kube_config()
    return clientt.CoreV1Api()


@pytest.fixture
def k8s_apps_client():
    """Fixture for Kubernetes Apps API client."""
    config.load_kube_config()
    return clientt.AppsV1Api()


@pytest.fixture
def wait_for_job_to_be_ready():
    """Helper fixture to wait for a Nomad job to be ready."""
    def _wait(nomad_client, job_id, timeout=60):
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                job_info = nomad_client.get(f"job/{job_id}")
                if job_info.get("Status") == "running":
                    allocations = nomad_client.get(f"job/{job_id}/allocations")
                    if allocations and any(
                            alloc.get("ClientStatus") == "running" for alloc in allocations):
                        return True
            except BaseException:
                pass
            time.sleep(2)
        return False

    return _wait
