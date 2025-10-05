output "projID" {
  description = "project id"
  value       = var.project_id
}

output "projNo" {
  description = "project number"
  value       = var.project_number
}


output "gke" {
  description = "GKE cluster details"
  value = {
    name           = google_container_cluster.logcorek8s.name
    release_channel = google_container_cluster.logcorek8s.release_channel[0].channel
    network        = google_container_cluster.logcorek8s.network
    subnet         = google_container_cluster.logcorek8s.subnetwork
    nodepool = {
      machine_type    = google_container_node_pool.primary_pool.node_config[0].machine_type
      disk            = "${google_container_node_pool.primary_pool.node_config[0].disk_type}-${google_container_node_pool.primary_pool.node_config[0].disk_size_gb}gb"
      service_account = google_container_node_pool.primary_pool.node_config[0].service_account
    }
  }
}
