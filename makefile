
run:
	go run app/services/tasker/main.go | go run app/tooling/logfmt/main.go

# ======================================================================
# Building containers

TASKER_VERSION := 0.1.0 

WORKER_VERSION := 0.1.0

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
		-t worker-amd64:$(WORKER_VERSION) \
		--build-arg BUILD_REF=$(WORKER_VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

tasker-image-update:
	cd zarf/k8s/cluster/tasker-pod; kustomize edit set image tasker-image=jnkroeker/tasker-amd64:$(TASKER_VERSION)

worker-image-update:
	cd zarf/k8s/cluster/worker-pod; kustomize edit set image worker-image=worker-amd64:$(WORKER_VERSION)

# ======================================================================
# Load and Run in k8s

k8s-apply:
	kustomize build zarf/k8s/cluster/tasker-pod | kubectl apply -f -
