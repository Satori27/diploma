.PHONY: open-grafana
open-grafana:
	kubectl --namespace monitoring get secrets kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo
	kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring

.PHONY: open-prom
open-prom:
	kubectl port-forward svc/kube-prometheus-stack-prometheus 9090:9090 -n monitoring

.PHONY: open-port-postgres
open-port-postgres:
	kubectl port-forward svc/postgres 5432:5432 -n database