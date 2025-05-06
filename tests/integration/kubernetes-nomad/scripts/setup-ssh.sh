#!/bin/bash
set -e

# Create SSH key for container communication
ssh-keygen -t rsa -f config/.ssh/id_rsa -N ""
cat config/.ssh/id_rsa.pub > config/.ssh/authorized_keys
chmod 600 config/.ssh/authorized_keys
chmod 600 config/.ssh/id_rsa