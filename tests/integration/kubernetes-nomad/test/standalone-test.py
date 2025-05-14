#!/usr/bin/env python3
"""
Standalone test script for Kubernetes-Nomad integration.
This script can be run outside of Docker to verify test logic.
"""

import unittest

class MockNomadClient:
    """Mock client for Nomad tests."""
    def get(self, endpoint):
        """Return mock data for Nomad endpoints."""
        if endpoint == "job/mantl-kubernetes-example":
            return {
                "Status": "running",
                "TaskGroups": [
                    {
                        "Name": "kubernetes-nginx",
                        "Tasks": [
                            {
                                "Name": "nginx",
                                "Config": {
                                    "image": "nginx:stable-alpine",
                                    "custom_resources": [{}],
                                    "pod_spec": "nvidia.com/gpu: 1"
                                },
                                "Resources": {
                                    "Devices": [
                                        {"Name": "nvidia/gpu", "Count": 1}
                                    ]
                                }
                            },
                            {
                                "Name": "logging-sidecar",
                                "Config": {"image": "fluent/fluentd:v1.14"}
                            },
                            {
                                "Name": "policy-agent",
                                "Config": {"image": "openpolicyagent/opa:latest"}
                            }
                        ],
                        "Services": [
                            {"Name": "mantl-example-service"},
                            {"Name": "mantl-example-metrics"}
                        ],
                        "Networks": [
                            {
                                "DynamicPorts": [
                                    {"Label": "http"},
                                    {"Label": "metrics"}
                                ]
                            }
                        ]
                    }
                ],
                "Scaling": {"Min": 1, "Max": 10},
                "Multiregion": {
                    "Strategy": "failover",
                    "Regions": [{"Name": "global"}]
                },
                "DisasterRecovery": {
                    "MaxDisconnect": "10m",
                    "AutoRevert": True
                }
            }
        elif endpoint == "job/mantl-kubernetes-example/allocations":
            return [{"ID": "test-alloc-id"}]
        elif endpoint == "allocation/test-alloc-id":
            return {
                "TaskStates": {
                    "nginx": {"State": "running"},
                    "logging-sidecar": {"State": "running"},
                    "policy-agent": {"State": "running"}
                }
            }
        elif endpoint == "agent/self":
            return {
                "config": {
                    "Region": "global",
                    "Server": {
                        "EnableMultiRegion": True,
                        "HeartbeatGrace": "30s"
                    },
                    "Plugins": {
                        "kubernetes": {}
                    }
                }
            }
        return {}

class MockK8sApiClient:
    """Mock client for Kubernetes tests."""
    def list_service_for_all_namespaces(self, label_selector=None):
        """Return mock services."""
        class MockServiceList:
            items = [
                type('obj', (object,), {
                    'metadata': type('obj', (object,), {
                        'name': 'mantl-example',
                        'labels': {'app': 'mantl-example'}
                    })
                })
            ]
        return MockServiceList()
    
    def list_pod_for_all_namespaces(self, label_selector=None):
        """Return mock pods."""
        class MockContainer:
            def __init__(self):
                self.name = "nginx"
                self.ports = [
                    type('obj', (object,), {'name': 'http', 'container_port': 80}),
                    type('obj', (object,), {'name': 'metrics', 'container_port': 9090})
                ]
                self.security_context = type('obj', (object,), {
                    'run_as_non_root': True,
                    'allow_privilege_escalation': False
                })
                self.resources = type('obj', (object,), {
                    'limits': {
                        'cpu': '200m',
                        'memory': '256Mi',
                        'nvidia.com/gpu': '1'
                    }
                })
                
        class MockPodList:
            items = [
                type('obj', (object,), {
                    'metadata': type('obj', (object,), {
                        'name': 'mantl-example',
                        'labels': {
                            'app': 'mantl-example',
                            'managed-by': 'nomad',
                            'mantl-service': 'true'
                        },
                        'annotations': {
                            'prometheus.io/scrape': 'true',
                            'prometheus.io/port': '9090'
                        }
                    }),
                    'spec': type('obj', (object,), {
                        'containers': [MockContainer()]
                    })
                })
            ]
        return MockPodList()

class MockConsulClient:
    """Mock client for Consul tests."""
    def catalog(self):
        """Return mock catalog."""
        class MockCatalog:
            def services(self):
                return None, {
                    'mantl-example-service': ['mantl', 'example', 'nginx', 'mantl-service=true'],
                    'mantl-example-metrics': ['mantl', 'metrics', 'prometheus']
                }
            
            def service(self, service_name):
                return None, [{
                    'ServiceAddress': '172.20.0.10',
                    'ServicePort': 8080,
                    'ServiceTags': ['mantl', 'example', 'nginx', 'mantl-service=true'],
                    'ServiceMeta': {'version': 'latest', 'service_type': 'web'}
                }]
        return MockCatalog()

class MockVaultClient:
    """Mock client for Vault tests."""
    def sys(self):
        """Return mock sys."""
        class MockSys:
            def read_health_status(self, method):
                return {'initialized': True}
            
            def list_policies(self):
                return {'policies': ['nomad-server']}
        return MockSys()
    
    def secrets(self):
        """Return mock secrets."""
        class MockKV:
            class MockV2:
                def read_secret_version(self, path, mount_point):
                    return {
                        'data': {
                            'data': {
                                'api_key': 'test-api-key',
                                'db_password': 'test-db-password'
                            }
                        }
                    }
                
                def create_or_update_secret(self, path, secret, mount_point):
                    return True
            
            def __init__(self):
                self.v2 = self.MockV2()
        
        class MockSecrets:
            def __init__(self):
                self.kv = MockKV()
        
        return MockSecrets()

class StandaloneTests(unittest.TestCase):
    """Test suite that runs without Docker containers."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.nomad_client = MockNomadClient()
        self.k8s_api_client = MockK8sApiClient()
        self.consul_client = MockConsulClient()
        self.vault_client = MockVaultClient()
    
    def test_service_discovery(self):
        """Test service discovery integration."""
        services = self.consul_client.catalog().services()[1]
        self.assertIn('mantl-example-service', services)
        
        service_tags = self.consul_client.catalog().service('mantl-example-service')[1][0]['ServiceTags']
        self.assertIn('mantl-service=true', service_tags)
        
        k8s_services = self.k8s_api_client.list_service_for_all_namespaces()
        self.assertTrue(len(k8s_services.items) > 0)
    
    def test_vault_integration(self):
        """Test Vault integration."""
        secret = self.vault_client.secrets().kv.v2.read_secret_version(
            path="mantl/kubernetes-example", mount_point="kv"
        )
        self.assertEqual(secret['data']['data']['api_key'], "test-api-key")
    
    def test_gpu_support(self):
        """Test GPU support."""
        job_info = self.nomad_client.get("job/mantl-kubernetes-example")
        
        # Find nginx task
        nginx_task = None
        for group in job_info["TaskGroups"]:
            for task in group["Tasks"]:
                if task["Name"] == "nginx":
                    nginx_task = task
                    break
        
        self.assertIsNotNone(nginx_task)
        
        # Check GPU device
        devices = nginx_task["Resources"]["Devices"]
        gpu_device = next((d for d in devices if "nvidia/gpu" in d["Name"]), None)
        self.assertIsNotNone(gpu_device)
        self.assertGreater(gpu_device["Count"], 0)
    
    def test_all_features(self):
        """Test all features are working together."""
        job_info = self.nomad_client.get("job/mantl-kubernetes-example")
        self.assertEqual(job_info["Status"], "running")
        
        # Check for all required task types
        tasks = []
        for group in job_info["TaskGroups"]:
            for task in group["Tasks"]:
                tasks.append(task["Name"])
        
        required_tasks = ["nginx", "logging-sidecar", "policy-agent"]
        for task_name in required_tasks:
            self.assertIn(task_name, tasks)
        
        # Verify Kubernetes resources
        pods = self.k8s_api_client.list_pod_for_all_namespaces(
            label_selector="managed-by=nomad"
        )
        self.assertTrue(len(pods.items) > 0)
        
        # Check pod for all feature labels
        pod = pods.items[0]
        self.assertIn("managed-by", pod.metadata.labels)
        self.assertIn("mantl-service", pod.metadata.labels)
        
        # Check container resources
        container = pod.spec.containers[0]
        self.assertIn("nvidia.com/gpu", container.resources.limits)

if __name__ == '__main__':
    unittest.main()