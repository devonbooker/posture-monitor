terraform {
  required_version = ">= 1.6"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.48"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.50"
    }
  }

  # Remote state on Cloudflare R2 (S3-compatible).
  # Access key + secret come from env:
  #   AWS_ACCESS_KEY_ID
  #   AWS_SECRET_ACCESS_KEY
  # The account ID in the endpoint is public (it appears in every R2 URL);
  # the secret that protects state is the R2 API token, which is gitignored.
  # use_lockfile uses S3 conditional writes (supported by R2) to prevent
  # concurrent applies without needing DynamoDB.
  backend "s3" {
    bucket = "posture-monitor-tfstate"
    key    = "terraform.tfstate"
    region = "auto"

    endpoints = {
      s3 = "https://6bac02fc6b2514598c9dfff1415a3737.r2.cloudflarestorage.com"
    }

    skip_credentials_validation = true
    skip_region_validation      = true
    skip_metadata_api_check     = true
    skip_requesting_account_id  = true
    skip_s3_checksum            = true
    use_lockfile                = true
  }
}
