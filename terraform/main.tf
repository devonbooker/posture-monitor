# Provider tokens are read from env vars:
#   HCLOUD_TOKEN            - Hetzner Cloud API (Read & Write)
#   CLOUDFLARE_API_TOKEN    - Cloudflare (Zone:DNS:Edit on devonbooker.dev)
# See terraform/.env.local.example for the shape of that file.
provider "hcloud" {}
provider "cloudflare" {}
