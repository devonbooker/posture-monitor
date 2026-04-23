output "server_ipv4" {
  value       = hcloud_server.app.ipv4_address
  description = "Public IPv4 of the app server (reachable from the internet)"
}

output "server_ipv6" {
  value       = hcloud_server.app.ipv6_address
  description = "Public IPv6 of the app server"
}

output "server_id" {
  value       = hcloud_server.app.id
  description = "Hetzner Cloud server ID"
}

output "posture_url" {
  value = "https://posture.${var.domain}"
}

output "grafana_url" {
  value = "https://grafana.${var.domain}"
}
