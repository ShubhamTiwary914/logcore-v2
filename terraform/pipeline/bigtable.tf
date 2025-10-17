resource "google_bigtable_instance" "logcorebt" {
  name = "logcorebt"
  deletion_protection = false
  labels = {
    "env": "dev"
  }
  cluster {
    cluster_id = "logcore-bt-instance"
    storage_type = "HDD"
    zone = "${var.region}-a" 
    num_nodes = 1 
  }
}

resource "google_bigtable_table" "iotsink" {
  name          = "iotsink"
  instance_name = google_bigtable_instance.logcorebt.name

#full payload
  column_family {
    family = "cf1"
  }
  #relation
  column_family {
    family = "cf2"  
  }
}