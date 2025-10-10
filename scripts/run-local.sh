## run and setup everything locally
## --------------------------------


RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' 

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

#kubernetes namespaces create (if not exists)
ns="verne"
create_k8s_ns() {
    local ns=$1
    if kubectl get ns "$ns" >/dev/null 2>&1; then
        log_skip "Namespace '$ns' already exists"
    else
        log_info "Creating namespace '$ns'..."
        kubectl create ns "$ns"
    fi
}
log_info "Setting up Kubernetes namespaces..."
create_k8s_ns "verne"
create_k8s_ns "observe"


# Pull Docker containers
log_info "Pulling Docker containers..."
docker pull sardinesszsz/k6-verne:latest
docker pull sardinesszsz/verne-listener:latest


# Add Helm repositories
log_info "Adding Helm repositories..."
helm repo add verne https://raw.githubusercontent.com/ShubhamTiwary914/logcore-v2/charts/ 2>/dev/null || log_skip "Verne repo already added"
helm repo add grafana https://grafana.github.io/helm-charts 2>/dev/null || log_skip "Grafana repo already added"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || log_skip "Prometheus repo already added"
helm repo update


# setup GCP emulators
log_info "Setting up GCP emulators..."
GCP_DIR="../services/gcp"

# Check if containers are running on ports 8085 and 8086
if lsof -i :8085 >/dev/null 2>&1 && lsof -i :8086 >/dev/null 2>&1; then
    # Ports are in use, check if containers are running or stopped
    CONTAINER_8085=$(docker ps -a --filter "publish=8085" --format "{{.Status}}")
    CONTAINER_8086=$(docker ps -a --filter "publish=8086" --format "{{.Status}}")
    
    if echo "$CONTAINER_8085" | grep -q "Up" && echo "$CONTAINER_8086" | grep -q "Up"; then
        log_skip "GCP emulators already running"
    else
        log_error "Containers on ports 8085/8086 exist but are stopped/exited."
        log_error "Remove them with: docker rm -f \$(docker ps -a --filter 'publish=8085' --filter 'publish=8086' -q)"
        exit 1
    fi
else
    log_info "Starting GCP emulators with docker-compose..."
    docker-compose -f "$GCP_DIR/docker-compose.yaml" up -d
    sleep 5
fi


log_info "Creating PubSub topic..."
bash "$GCP_DIR/pubsub/create-topic.sh" || log_skip "Topic may already exist"

log_info "Initializing BigTable..."
if command -v conda &> /dev/null; then
    conda run -n logcore python "$GCP_DIR/bigtable/init.py" || log_error "Failed to initialize BigTable"
else
    log_error "Conda not found. Please run manually: conda run -n logcore python $GCP_DIR/bigtable/init.py"
fi

#verne setup
VERNE_DIR="../k8s/dev/verne"
log_info "Installing Verne chart..."
if helm list -n verne | grep -q "verne"; then
    log_skip "Verne chart already installed"
else
    helm install verne verne/verne -n verne -f "$VERNE_DIR/dev-values.yaml"
fi

#setup observability stack
OBSERVE_DIR="../k8s/dev/observe"
log_info "Setting up observability stack..."

#loki
if helm list -n observe | grep -q "loki"; then
    log_skip "Loki chart already installed"
else
    log_info "Installing Loki..."
    helm install loki grafana/loki -n observe -f "$OBSERVE_DIR/values/loki.yaml"
fi

#prom
if helm list -n observe | grep -q "prometheus"; then
    log_skip "Prometheus chart already installed"
else
    log_info "Installing Prometheus..."
    helm install prometheus prometheus-community/prometheus -n observe -f "$OBSERVE_DIR/values/prom.yaml"
fi

# grafana dashboards config
if kubectl get configmap grafana-dashboards -n observe >/dev/null 2>&1; then
    log_skip "Grafana dashboards configmap already exists"
else
    log_info "Creating Grafana dashboards configmap..."
    kubectl create configmap grafana-dashboards -n observe \
      --from-file=listener-logs.json=$OBSERVE_DIR/graf-exports/listener-logs.json \
      --from-file=verne-metrics.json=$OBSERVE_DIR/graf-exports/verne-metrics.json
fi

#graf setup
if helm list -n observe | grep -q "grafana"; then
    log_skip "Grafana chart already installed"
else
    log_info "Installing Grafana..."
    helm install grafana grafana/grafana -n observe -f "$OBSERVE_DIR/values/graf.yaml"
fi

#alloy setup
if helm list -n observe | grep -q "alloy"; then
    log_skip "Alloy chart already installed"
else
    log_info "Installing Alloy..."
    helm install alloy grafana/alloy -n observe -f "$OBSERVE_DIR/values/alloy.yaml"
fi


log_info "Setup complete!"
GRAFANA_IP=$(kubectl get svc grafana -n observe -o jsonpath='{.spec.clusterIP}')
GRAFANA_PORT=3045
GRAFANA_PASS=$(kubectl get secret grafana -n observe -o jsonpath="{.data.admin-password}" | base64 --decode)

log_info "Grafana Dashboard URL: http://${GRAFANA_IP}:${GRAFANA_PORT}"
log_info "Username: admin"
log_info "Password: ${GRAFANA_PASS}"