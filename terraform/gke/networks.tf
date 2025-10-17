resource "google_compute_network" "logcore-net" {
    project = var.project_id
    name = "logcore-net"
    auto_create_subnetworks = false
    mtu = 1460
}

resource "google_compute_subnetwork" "logcore-lowa-subnet" {
    name = "logcore-lowa-subnet"
    network = google_compute_network.logcore-net.id
    region = var.region
    ip_cidr_range = "10.10.0.0/24"
    secondary_ip_range {
        range_name    = "logcore-loaw-subnet-range"
        ip_cidr_range = "10.10.10.0/24" //252 host IPs
    }
}