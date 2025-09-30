GRAFIP=$(kubectl get svc grafana -n observe -o jsonpath='{.spec.clusterIP}')
GRAFPASS=$(kubectl get secret --namespace observe grafana -o jsonpath="{.data.admin-password}" | base64 --decode)

echo "Grafana Web UI for k8s accessible at: http://$GRAFIP:80"
echo "Username: admin"
echo "Pass: $GRAFPASS"