resource "google_container_cluster" "default" {
    name     = "logcorek8s"
    location = var.region
    enable_autopilot = true
    network = google_compute_network.logcore-net.id
    subnetwork = google_compute_subnetwork.logcore-lowa-subnet.id
    project = var.project_id
    deletion_protection = false

    cluster_autoscaling {
        auto_provisioning_defaults {
        service_account = var.logcore_serviceacc
        }
    }
    release_channel {
        channel = "REGULAR"
    }
    
    resource_labels = {
        service = "logcorek8s"
    }

    control_plane_endpoints_config {
        ip_endpoints_config {
          enabled = false
        }
        dns_endpoint_config {
          allow_external_traffic = true
        }
    }

    private_cluster_config {
      enable_private_nodes = false
      enable_private_endpoint = false
    }

    logging_service    = "logging.googleapis.com/kubernetes"
    monitoring_service = "monitoring.googleapis.com/kubernetes"
}