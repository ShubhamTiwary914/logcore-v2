#run k6 load test on the prod GKE verneMQTT pods
#usage: ./k6-prodrun.sh [duration] [vus]
# defaults: duration=10, vus=5

GREEN='\033[0;32m'
NC='\033[0m' 
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

DURATION="${1:-10}"
VUS="${2:-5}"

log_info "Running K6 load test with conditions:"
log_info "Duration: ${DURATION}s"
log_info "Virtual Users: ${VUS}"
log_info "Interval: 1"

VERNE_IP=$(kubectl get svc -n verne vernemq-broker -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
log_info "Running K6 Load on verneIP (external LB): ${VERNE_IP}"

docker run --rm --network host \
  -e RELATION=machine_metrics\
  -e DURATION="$DURATION" \
  -e INTERVAL=1 \
  -e VUS="$VUS" \
  -e VERNE_IP="$VERNE_IP" \
  sardinesszsz/k6-verne:latest