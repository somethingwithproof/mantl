#!/bin/bash
# Script to fix serious Python linting issues that weren't addressed by autopep8

set -e

echo "Fixing serious Python linting issues..."

# Fix various syntax errors and undefined names
# These are the most critical fixes to make code run properly

# Fix roles/collectd/files/mesos-agent.py
echo "Fixing roles/collectd/files/mesos-agent.py..."
sed -i '' 's/except urllib2.URLError, e:/except (urllib2.URLError, socket.error) as e:/g' roles/collectd/files/mesos-agent.py

# Fix roles/collectd/files/mesos-master.py
echo "Fixing roles/collectd/files/mesos-master.py..."
sed -i '' 's/except urllib2.URLError, e:/except (urllib2.URLError, socket.error) as e:/g' roles/collectd/files/mesos-master.py

# Fix roles/collectd/files/zookeeper-collectd-plugin.py
echo "Fixing roles/collectd/files/zookeeper-collectd-plugin.py..."
sed -i '' 's/except socket.error, e:/except (socket.error, IOError) as e:/g' roles/collectd/files/zookeeper-collectd-plugin.py

# Fix roles/lvm/files/mantl-storage-setup.py print statement for Python 3
echo "Fixing roles/lvm/files/mantl-storage-setup.py..."
sed -i '' 's/print "/print("/g' roles/lvm/files/mantl-storage-setup.py

# Fix import of defaultdict in plugins/inventory/terraform.py
echo "Fixing plugins/inventory/terraform.py..."
sed -i '' 's/from collections import defaultdic/from collections import defaultdict/g' plugins/inventory/terraform.py
sed -i '' 's/subne/subnet/g' plugins/inventory/terraform.py

# Fix import in testing/travis.py
echo "Fixing testing/travis.py..."
sed -i '' 's/import socke/import socket/g' testing/travis.py

# Fix the import issue in roles/collectd/files/marathon-collectd-plugin.py
echo "Fixing roles/collectd/files/marathon-collectd-plugin.py..."
# Replace basestring with str for Python 3 compatibility
sed -i '' 's/isinstance(value, basestring)/isinstance(value, str)/g' roles/collectd/files/marathon-collectd-plugin.py

# Fix the undefined pytest in test files
echo "Fixing pytest imports in test files..."
for file in $(grep -l "pytes" --include="*.py" -r tests/); do
  sed -i '' 's/import pytes/import pytest/g' "$file"
done

# Fix the undefined client in test/standalone-test.py
echo "Fixing test/standalone-test.py..."
sed -i '' 's/import unittes/import unittest/g' tests/integration/kubernetes-nomad/test/standalone-test.py

# Fix other pytest-related issues in test files
echo "Fixing client imports in test_crd.py..."
sed -i '' 's/import kubernetes.clien/import kubernetes.client/g' tests/integration/kubernetes-nomad/test/test_crd.py

# Fix conftest.py
echo "Fixing conftest.py..."
sed -i '' 's/return clien/return client/g' tests/integration/kubernetes-nomad/test/conftest.py
sed -i '' 's/return _wai/return _wait/g' tests/integration/kubernetes-nomad/test/conftest.py

# Fix missing import in library/kube.py for AnsibleModule
echo "Fixing kube.py..."
sed -i '' 's/from ansible.module_utils.basic import \*/from ansible.module_utils.basic import AnsibleModule/g' library/kube.py

echo "Most serious linting issues have been fixed. Some issues may still require manual fixes."