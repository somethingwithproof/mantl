#!/bin/bash
set -e

# Run Kubernetes-Nomad integration tests on GCP
echo "Setting up GCP environment for Kubernetes-Nomad integration tests..."

# Get GCP configuration
GCP_PROJECT=${GCP_PROJECT:-mantl-project}
GCP_ZONE=${GCP_ZONE:-us-central1-a}
GCP_MACHINE_TYPE=${GCP_MACHINE_TYPE:-n1-standard-4}

echo "Using GCP project: $GCP_PROJECT"
echo "Using GCP zone: $GCP_ZONE"
echo "Using machine type: $GCP_MACHINE_TYPE"

# Check gcloud CLI
if ! command -v gcloud &> /dev/null; then
    echo "gcloud CLI not found. Please install it first."
    exit 1
fi

# Create GCS bucket
echo "Setting up GCS bucket..."
GCS_BUCKET="mantl-integration-tests"
GCS_PREFIX="kubernetes-nomad-$(date +%Y%m%d%H%M%S)"

gsutil mb -p $GCP_PROJECT gs://$GCS_BUCKET || true

# Upload test files
echo "Uploading test files to GCS..."
gsutil -m cp -r .. gs://$GCS_BUCKET/$GCS_PREFIX

# Create startup script
cat > startup-script.sh <<'EOF'
#!/bin/bash
apt-get update
apt-get install -y docker.io docker-compose
systemctl enable docker
systemctl start docker
mkdir -p /opt/mantl-tests
gsutil -m cp -r gs://BUCKET_NAME/PREFIX/* /opt/mantl-tests/
cd /opt/mantl-tests
chmod +x run-tests.sh
./run-tests.sh | tee /opt/mantl-tests/results/test-output.log
gsutil -m cp -r /opt/mantl-tests/results/* gs://BUCKET_NAME/PREFIX/results/
shutdown -h now
EOF

# Replace placeholders in script
sed -i "s/BUCKET_NAME/$GCS_BUCKET/g" startup-script.sh
sed -i "s/PREFIX/$GCS_PREFIX/g" startup-script.sh

# Upload startup script
gsutil cp startup-script.sh gs://$GCS_BUCKET/$GCS_PREFIX/startup-script.sh

# Launch GCE instance
echo "Launching GCE instance..."
INSTANCE_NAME="mantl-k8s-nomad-test-$(date +%Y%m%d%H%M%S)"

gcloud compute instances create $INSTANCE_NAME \
    --project=$GCP_PROJECT \
    --zone=$GCP_ZONE \
    --machine-type=$GCP_MACHINE_TYPE \
    --subnet=default \
    --network-tier=PREMIUM \
    --maintenance-policy=MIGRATE \
    --service-account=default \
    --scopes=https://www.googleapis.com/auth/devstorage.read_write,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append \
    --tags=http-server,https-server \
    --create-disk=auto-delete=yes,boot=yes,device-name=instance-1,image=projects/debian-cloud/global/images/debian-10-buster-v20220822,mode=rw,size=50,type=pd-balanced \
    --metadata=startup-script-url=gs://$GCS_BUCKET/$GCS_PREFIX/startup-script.sh \
    --no-shielded-secure-boot \
    --shielded-vtpm \
    --shielded-integrity-monitoring \
    --reservation-affinity=any

echo "Instance launched: $INSTANCE_NAME"
echo "Waiting for tests to complete..."

# Wait for instance to terminate (this script will poll the instance status)
while true; do
    STATUS=$(gcloud compute instances describe $INSTANCE_NAME --zone=$GCP_ZONE --format="value(status)")
    if [ "$STATUS" != "RUNNING" ]; then
        echo "Instance terminated or in an unexpected state: $STATUS"
        break
    fi
    echo "Instance still running, waiting..."
    sleep 60
done

# Download results
echo "Downloading test results..."
mkdir -p ../results-gcp
gsutil -m cp -r gs://$GCS_BUCKET/$GCS_PREFIX/results/* ../results-gcp/

echo "Tests completed. Results available in: ../results-gcp"