resource "google_pubsub_topic" "mqtt-source-topic" {
  name = "mqtt-source"
}

resource "google_pubsub_topic_iam_member" "job_complete_publisher" {
  topic = google_pubsub_topic.mqtt-source-topic.name
  role  = "roles/pubsub.publisher"
  member = "serviceAccount:${var.logcore_serviceacc}"
}