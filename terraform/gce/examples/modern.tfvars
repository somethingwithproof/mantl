# Example tfvars for GCE modern paths
# Network (apply with PR #46)
use_modern_gce_network       = true
# After applying network_modern, set this to the output value
modern_subnetwork_self_link  = "projects/your-proj/regions/us-central1/subnetworks/mantl-subnet"

# Instances (apply with PR #45)
use_modern_gce_schema = true
allowed_cidrs         = ["203.0.113.0/24"]

gce_boot_image_family  = "centos-7"
gce_boot_image_project = "centos-cloud"
# Optional CMEK
#gce_boot_kms_key = "projects/your-proj/locations/us/keyRings/kr/cryptoKeys/key"

gce_public_ip = false

# Legacy counts should be zeroed when modern is enabled
control_count    = 3
worker_count     = 2
kubeworker_count = 0
