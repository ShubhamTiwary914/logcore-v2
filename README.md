
## Table of Contents

- [About](https://github.com/ShubhamTiwary914/logcore-v2#about)
- [Architecture](https://github.com/ShubhamTiwary914/logcore-v2#architecture)
  - [Rough Workings](https://github.com/ShubhamTiwary914/logcore-v2#rough-overview-of-the-tools-and-workings)
  - [Dev vs Prod Environments](https://github.com/ShubhamTiwary914/logcore-v2#difference-between-the-local-version-and-on-gcp-premise)
-  [Local Setup](https://github.com/ShubhamTiwary914/logcore-v2#setting-up-locally)
-  [Checklist](https://github.com/ShubhamTiwary914/logcore-v2#things-to-add)


---

# About

A IOT streaming platform to simulate lots of devices (10k+ concurrent connections) and make out the flow with dashboards. <br />
Earlier version was [logcore-v1](https://github.com/ShubhamTiwary914/logcore), which worked  on top of Docker swarm on Virtual Machines locally.

> So this has version additional support for:
- Using a MQTT cluster (vernMQ) to handle much more load of traffic and concurrent connections
- Load Testing with K6
- Migrating from local homelab to GCP's GKE(Kubernetes Engine) to scale better
- Earlier, TimescaleDB was used, now replaced with BigTable
- Support to run locally as well (with K3s and GCP local emulators)
- Helm charts configs for easy Kubernetes setup 

---

# Architecture
<img width="1388" height="529" alt="screenshot_2025-10-01-172856" src="https://github.com/user-attachments/assets/63a30adf-8d4f-4acf-99bb-951be834d306" />

<br />

### Rough overview of the tools and workings:
- K6: devices mock for load testing (send MQTT messages to message broker - verne)
- Kubernetes cluster has the following namespaces:
  - Observe:  observability stack with prometheus for metrics, alloy & loki for logs, grafana for visualization
  - Verne:  MQTT pod for message broker, and a listener that supplies messages from MQTT -> PubSub
- PubSub: Holds messages in source/ topic
- Dataflow: [streaming ETL pipeline](https://hazelcast.com/foundations/event-driven-architecture/streaming-etl/) to transform & push data to data storages
- BigTable:  holds the full data, really scalable, no schema needed
- FireStore: device shadows (active devices list)
- NextJS Dashboard:  add new type, user side views


### Difference between the local version and on GCP premise:
- [K3s](https://k3s.io/) used in place of K8s(Kubernetes), its more lightweight and works for local nodes (even in lightweight devices like Pi)
- Dataflow replaced with DirectRunner in [Apache Beam SDK](https://beam.apache.org/about/) (internally it's the same thing)
- [GCP emulators (beta)](https://cloud.google.com/sdk/gcloud/reference/beta/emulators) for pubsub & bigtable (internally the same APIs)
 

---


# Setting up locally

Tools needed to run this locally:
- [K3s](https://docs.k3s.io/installation)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Taskfile](https://taskfile.dev/docs/installation)
- [WSL](https://ubuntu.com/desktop/wsl) (only for windows)
- [Helm](https://helm.sh/docs/intro/install/)
- [Docker](https://docs.docker.com/desktop/) (with compose support- recommended: get the Desktop version, CLI works too but seperate)

<br />

1. Clone repo
```bash
git clone https://github.com/ShubhamTiwary914/logcore-v2.git
cd logcore-v2
```

2. Run setup script
```bash
task _run-local
```

Preview of setup logs:
```bash
task: [_run-local] bash -c "./run-local.sh"
[INFO] Setting up Kubernetes namespaces...
[INFO] Creating namespace 'verne'...
namespace/verne created
[INFO] Creating namespace 'observe'...
namespace/observe created
.
.
[INFO] Setup complete!
[INFO] Grafana Dashboard URL: http://10.43.120.70:3045
[INFO] Username: admin
[INFO] Password: *******
```

Use this URL on browser to view the Grafana Dashboard with the credentials

> [!NOTE]
> Grafana dashboard may take a few secs to setup, you can check if it ready via: `kubectl get pods -n observe | grep grafana`, which shows:
> ```bash
> grafana-66bd6889cf-b4f7t   1/1     Running   0    2m10s
> ```

The dashboard section has custom metrics being shown, previews:
<img width="1557" height="961" alt="screenshot_2025-10-01-174858" src="https://github.com/user-attachments/assets/6a001d11-6505-49aa-b482-00f97d43b783" />
<img width="1049" height="744" alt="screenshot_2025-10-01-180430" src="https://github.com/user-attachments/assets/4c4cc03c-eaac-4cd8-b236-9ecc0d80f350" />


And in case you forget/remove the logs, getting back the address & creds. to access grafana:
```bash
task observe-grafana-access
```


---

# Things to add:
- [ ] Firestore support for device shadows
- [ ] NextJS dashboard to add additional schemas and user side data trace
- [ ] Migrating the cluster from K3s and GCP emulators to GCP
- [ ] Benchmarking comparison between local nodes & GKE after final deployments


