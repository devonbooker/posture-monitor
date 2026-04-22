resource "hcloud_ssh_key" "main" {
  name       = var.ssh_key_name
  public_key = var.ssh_public_key

  labels = {
    managed_by = "terraform"
    project    = "posture-monitor"
  }
}

resource "hcloud_server" "app" {
  name        = var.server_name
  server_type = var.server_type
  location    = var.server_location
  image       = var.server_image
  ssh_keys    = [hcloud_ssh_key.main.name]
  user_data   = file("${path.module}/cloud-init.yaml")

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }

  labels = {
    managed_by = "terraform"
    project    = "posture-monitor"
    role       = "app"
  }

  lifecycle {
    # image, ssh_keys, and user_data are only consumed at create-time by Hetzner.
    # The imported server has no ssh_keys/user_data attached (password-bootstrapped,
    # keys pushed manually afterward), so ignoring them is the only way to keep
    # the existing VM in state without forcing a destroy+recreate cycle that
    # would mean cert re-issuance and SQLite data loss. If the VM ever dies and
    # Terraform recreates it, these values WILL take effect on the fresh instance.
    # public_net is ignored because any diff there makes the Hetzner provider
    # reconfigure networking and briefly bounce the VM (learned the hard way
    # during the import apply).
    ignore_changes = [
      ssh_keys,
      user_data,
      image,
      public_net,
    ]
  }
}
