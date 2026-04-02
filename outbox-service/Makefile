.PHONY: gen-sqlc
gen-sqlc:
	sqlc generate -f internal/db/sqlc.yaml

.PHONY: gen-migrations
gen-migrations:
	goose -dir migrations create $(NAME) sql

.PHONY: run-skaffold
run-skaffold:
	sudo KUBECONFIG=/home/satori/.kube/config skaffold run --kube-context=kubernetes-admin@cluster.local