# khyme
Chyme ETL refactored into services for running Kubernetes

# Changelog

01-09-2023

    Very important early on to nail down Logging, Configuration, Error handling, Build and Deployment process.

    I want a working, debuggable, maintainable app every step of the way to completion.

01-31-2023

    After a fix:
    * update the service version in makefile
    * run `make <service name>` to generate a new image
    * push image to Docker Hub with `docker push <local image name:tag>`
    * update image version to be used by k8s deployment by executing `make <service>-image-update`
    * delete existing deployment (if there is one)
    * run `make k8s-apply` to start a new deployment with the updated image

02-02-2023

    https://www.weave.works/blog/kubectl-port-forward

    K8s Port-Forwarding using kubectl allows me to access my cluster from a local browser for simple debugging

    `kubectl port-forward <pod name> <local port>:<service port>` (add --namespace=khyme-system if namespace not configured)

    find the Khyme debug endpoint in a local browser at http://localhost:4000/debug/pprof/

