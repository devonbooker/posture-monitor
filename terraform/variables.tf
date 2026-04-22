variable "server_name" {
  type        = string
  default     = "posture-monitor"
  description = "Logical name of the Hetzner Cloud server"
}

variable "server_type" {
  type        = string
  default     = "cax11"
  description = "Hetzner Cloud server type (cax11 = ARM64, 2 vCPU, 4GB)"
}

variable "server_location" {
  type        = string
  default     = "nbg1"
  description = "Hetzner Cloud location (nbg1 = Nuremberg)"
}

variable "server_image" {
  type        = string
  default     = "ubuntu-22.04"
  description = "OS image used ONLY on fresh create. Imported server drifts from this value harmlessly."
}

variable "ssh_key_name" {
  type        = string
  default     = "hetzner-uptime-monitor"
  description = "Name of the SSH key in the Hetzner project. Matches the existing imported key."
}

variable "ssh_public_key" {
  type        = string
  default     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIELD88IGAaw7Sx+Szd7Il03Qc0ZaR5Qu/Wo7NfgGmGsq"
  description = "Public half of the Hetzner provisioning SSH key. Public keys are not secrets. Override via TF_VAR_ssh_public_key if rotating."
}

variable "cloudflare_zone_id" {
  type        = string
  default     = "7e4c17895476e58fbda8207fc230590a"
  description = "Cloudflare zone ID for devonbooker.dev"
}

variable "domain" {
  type        = string
  default     = "devonbooker.dev"
  description = "Root domain managed in Cloudflare"
}
