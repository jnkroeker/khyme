# khyme
Chyme ETL refactored into services for running Kubernetes

## Build a new image and redeploy after a bug fix or new feature

    * update the service version in makefile
    * run `make <service name>` to generate a new image
    * push image to Docker Hub with `docker push <local image name:tag>`
    * update image version to be used by k8s deployment by executing `make <service>-image-update`
    * delete existing deployment (if there is one)
    * run `make k8s-<service>-apply` to start a new deployment with the updated image

## Use kubectl port-forwarding to access cluster from local machine

    `kubectl port-forward <pod name> <local port>:<service port>` (add --namespace=khyme-system if namespace not configured)

# Changelog

01-09-2023

    Very important early on to nail down Logging, Configuration, Error handling, Build and Deployment process.

    I want a working, debuggable, maintainable app every step of the way to completion.

02-02-2023

    https://www.weave.works/blog/kubectl-port-forward

    K8s Port-Forwarding using kubectl allows me to access my cluster from a local browser for simple debugging

    `kubectl port-forward <pod name> <local port>:<service port>` (add --namespace=khyme-system if namespace not configured)

    find the Khyme debug endpoint in a local browser at http://localhost:4000/debug/pprof/

02-03-2023

    Tasker and Worker services v1 running in separate pods within same namespace on k8s cluster

