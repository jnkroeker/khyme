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
		-t tasker-amd64:$(TASKER_VERSION) \
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