provider "google" {
  project = var.project_id
  region = var.region
}

variable "project_id" {
  description = "GCP Project ID"
}

variable "region" {
  description = "GCP Project region"
}

variable "logcore_serviceacc" {
  description = "Service account to handle GCP logcore workloads - for GKE"
}