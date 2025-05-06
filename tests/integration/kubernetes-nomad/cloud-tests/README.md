# Cloud Provider Integration Tests

This directory contains scripts to run the Kubernetes-Nomad integration tests on various cloud providers.

## Supported Cloud Providers

- **AWS**: Amazon Web Services (EC2)
- **GCP**: Google Cloud Platform (GCE)

## Prerequisites

### AWS
- AWS CLI installed and configured with appropriate credentials
- The necessary permissions to create EC2 instances, S3 buckets, and related resources
- Update the security group and subnet IDs in the script

### GCP
- gcloud CLI installed and configured with appropriate credentials
- The necessary permissions to create GCE instances, GCS buckets, and related resources
- Update the project ID and other parameters as needed

## Running Tests

### AWS

```bash
# Run with default settings
./run-aws.sh

# Or with custom settings
AWS_REGION=us-east-1 AWS_INSTANCE_TYPE=m5.xlarge ./run-aws.sh
```

### GCP

```bash
# Run with default settings
./run-gcp.sh

# Or with custom settings
GCP_PROJECT=my-project GCP_ZONE=us-west1-b GCP_MACHINE_TYPE=n2-standard-4 ./run-gcp.sh
```

## Test Results

The test results will be downloaded to:

- AWS: `../results-aws/`
- GCP: `../results-gcp/`

Each run creates a timestamped directory to avoid overwriting previous results.

## Notes

- The tests will automatically shut down the instances after completion to minimize costs
- All test artifacts are stored in cloud storage (S3 or GCS) for review
- The scripts use temporary instances to avoid leaving long-running resources