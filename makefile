
# ======================================================================
# Define dependencies

GOLANG          := golang:1.20
ALPINE          := alpine:3.18
KIND            := kindest/node:v1.27.3
POSTGRES        := postgres:15.3
VAULT           := hashicorp/vault:1.13
GRAFANA         := grafana/grafana:9.5.3
PROMETHEUS      := prom/prometheus:v2.44.0
TEMPO           := grafana/tempo:2.1.1
TELEPRESENCE    := datawire/ambassador-telepresence-manager:2.14.0

# ======================================================================
# Install dependencies

dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest

dev-brew-common:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev-brew: dev-brew-common
	brew list datawire/blackbird/telepresence || brew install datawire/blackbird/telepresence

dev-docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(VAULT)
	docker pull $(GRAFANA)
	docker pull $(PROMETHEUS)
	docker pull $(TEMPO)
	docker pull $(TELEPRESENCE)

# ======================================================================
# Systems startup

run-tasker:
	go run app/services/tasker/main.go | go run app/tooling/logfmt/main.go

run-worker:
	go run app/services/worker/main.go --help | go run app/tooling/logfmt/main.go

seed:
	go run app/tooling/khyme-admin/main.go seed

vault-init:
	go run app/tooling/khyme-admin/main.go vault-init

# ======================================================================
# Testing running systems

# Database access
# kubectl port-forward <pod name> 5432:5432 --namespace=database-system
# dblab --host 0.0.0.0 --user postgres --db postgres --pass postgres --ssl disable --port 5432 --driver postgres

# ======================================================================
# Building containers

TASKER_VERSION := 0.2.2

WORKER_VERSION := 0.1.2

all: 
	tasker 
	worker

tasker:
	docker build \
		-f zarf/docker/dockerfile.tasker \
		-t jnkroeker/tasker-amd64:$(TASKER_VERSION) \
		--build-arg BUILD_REF=$(TASKER_VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

worker:
	docker build \
		-f zarf/docker/dockerfile.worker \
		-t jnkroeker/worker-amd64:$(WORKER_VERSION) \
		--build-arg BUILD_REF=$(WORKER_VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

tasker-image-update:
	cd zarf/k8s/cluster/tasker-pod; kustomize edit set image tasker-image=jnkroeker/tasker-amd64:$(TASKER_VERSION)

worker-image-update:
	cd zarf/k8s/cluster/worker-pod; kustomize edit set image worker-image=jnkroeker/worker-amd64:$(WORKER_VERSION)

# ======================================================================
# Load and Run in k8s

k8s-tasker-apply:
	kustomize build zarf/k8s/cluster/tasker-pod | kubectl apply -f -

k8s-worker-apply:
	kustomize build zarf/k8s/cluster/worker-pod | kubectl apply -f -

k8s-database-apply:
	kustomize build zarf/k8s/base/database-pod | kubectl apply -f -

k8s-vault-apply:
	kustomize build zarf/k8s/base/vault-pod | kubectl apply -f -

k8s-tasker-logs:
	kubectl logs -l app=tasker --namespace=khyme-system --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

k8s-worker-logs:
	kubectl logs -l app=worker --namespace=khyme-system --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

k8s-tasker-restart:
	kubectl rollout restart deployment tasker-pod --namespace=khyme-system
