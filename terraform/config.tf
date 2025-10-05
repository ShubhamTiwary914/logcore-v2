provider "google" {
  project = var.project_id
  region = var.region
}

#variables
variable "project_id" {
  description = "GCP Project ID"
}

variable "project_number" {
  description = "GCP Project Number"
}

variable "region" {
  description = "GCP Project region"
}