## Context

Config for the k8s services, includes:
- VerneMQ broker -> MQTT brokers
- Monitor -> Prometheus Operator with grafana (analyse pods load, etc)
- Listener ->  connects: (verneMQ <-> listener <-> pubsub) --->  side-container beside the verneMQ broker container in each pod

<hr>

### Dev Mode vs Prod


**Dev Mode**:
- k8s is running as k3s instead (local nodes)
- gcp emulator (for pubsub & bigtable) being used instead
- listener --> pushes to pubsub local

<br />


**Dev Mode**:
- k3s replaced by GKE (same config - cause of custom mode, not auto)
- actual GCP service, not the emulator (main difference: IAM - Service ACs)
- listener --> pushes to pubsub 
