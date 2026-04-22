resource "cloudflare_record" "posture" {
  zone_id = var.cloudflare_zone_id
  name    = "posture"
  type    = "A"
  content = hcloud_server.app.ipv4_address
  ttl     = 1 # 1 = Auto
  proxied = false
  comment = "Managed by Terraform - posture-monitor dashboard"
}

resource "cloudflare_record" "grafana" {
  zone_id = var.cloudflare_zone_id
  name    = "grafana"
  type    = "A"
  content = hcloud_server.app.ipv4_address
  ttl     = 1 # 1 = Auto
  proxied = false
  comment = "Managed by Terraform - posture-monitor grafana"
}
