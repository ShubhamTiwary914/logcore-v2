kubectl delete ns verne
kubectl delete ns observe

docker compose -f ./../services/gcp/docker-compose.yaml down