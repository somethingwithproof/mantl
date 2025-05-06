#!/bin/bash
set -e

# Run Kubernetes-Nomad integration tests on AWS
echo "Setting up AWS environment for Kubernetes-Nomad integration tests..."

# Get AWS credentials (assuming AWS CLI is configured)
AWS_REGION=${AWS_REGION:-us-west-2}
AWS_INSTANCE_TYPE=${AWS_INSTANCE_TYPE:-t3.xlarge}

echo "Using AWS region: $AWS_REGION"
echo "Using instance type: $AWS_INSTANCE_TYPE"

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo "AWS CLI not found. Please install it first."
    exit 1
fi

# Upload test files
echo "Uploading test files to S3..."
S3_BUCKET="mantl-integration-tests"
S3_PREFIX="kubernetes-nomad-$(date +%Y%m%d%H%M%S)"

aws s3 mb s3://$S3_BUCKET --region $AWS_REGION || true
aws s3 sync .. s3://$S3_BUCKET/$S3_PREFIX --exclude "cloud-tests/*" --exclude ".git/*"

# Launch EC2 instance
echo "Launching EC2 instance..."
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id ami-0c55b159cbfafe1f0 \
    --instance-type $AWS_INSTANCE_TYPE \
    --security-group-ids sg-12345 \
    --subnet-id subnet-12345 \
    --block-device-mappings 'DeviceName=/dev/sda1,Ebs={VolumeSize=50}' \
    --user-data "#!/bin/bash
apt-get update
apt-get install -y docker.io docker-compose awscli
systemctl enable docker
systemctl start docker
mkdir -p /opt/mantl-tests
aws s3 sync s3://$S3_BUCKET/$S3_PREFIX /opt/mantl-tests
cd /opt/mantl-tests
chmod +x run-tests.sh
./run-tests.sh | tee /opt/mantl-tests/results/test-output.log
aws s3 sync /opt/mantl-tests/results s3://$S3_BUCKET/$S3_PREFIX/results
shutdown -h now" \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=mantl-kubernetes-nomad-test}]' \
    --region $AWS_REGION \
    --query 'Instances[0].InstanceId' \
    --output text)

echo "Instance launched: $INSTANCE_ID"
echo "Waiting for tests to complete..."

# Wait for instance to terminate
aws ec2 wait instance-terminated --instance-ids $INSTANCE_ID --region $AWS_REGION

# Download results
echo "Downloading test results..."
mkdir -p ../results-aws
aws s3 sync s3://$S3_BUCKET/$S3_PREFIX/results ../results-aws

echo "Tests completed. Results available in: ../results-aws"