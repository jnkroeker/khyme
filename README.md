# khyme
Chyme ETL refactored into services for running Kubernetes

## Build a new service image and redeploy after a bug fix or new feature

    * update the service version in makefile
    * run `make <service name>` to generate a new image
    * push image to Docker Hub with `docker push <local image name:tag>`
    * update image version to be used by k8s deployment by executing `make <service>-image-update`
    * delete existing deployment (if there is one)
    * run `make k8s-<service>-apply` to start a new deployment with the updated image

## Deploy PostgreSQL database to Khyme cluster

    * run `make k8s-database-apply` to start new deployment in separate namespace to Tasker and Worker
    * use dblab (github.com/danvergara/dblab) to establish connection to postgresql database (see Makefile)

## Use kubectl port-forwarding to access cluster from local machine

    `kubectl port-forward <pod name> <local port>:<service port>` (add --namespace=<namespace> if namespace not configured)

## Seed Database with Test Data and Perform Test Queries

    * Three pods: Tasker, Worker and Database must be up on k8s cluster
    * Connections from local machine to Tasker and Database must be opened with below commands

        `kubectl port-forward <database pod name> 5432:5432 --namespace=database-system`
        `kubectl port-forward <tasker pod name> 3000:3000 --namespace=khyme-system`
    
    * Seed the database with `make khyme-admin` command
    * execute curl requests from terminal to test Create, Read, Destroy endpoints

        GET:  `curl http://localhost:3000/v1/tasks/1/1`
        POST: `curl http://localhost:3000/v1/tasks -H "Content-Type: text/plain" -d '"<url text string>"'`
        DEL:  `curl http://localhost:3000/v1/tasks/<task id>`

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

02-24-2023

    Tasker readiness debug endpoint works. Tasker accepts GET, POST, DELETE http requests so long as kubectl port forwarding
        to 3000 is open. 

    Existing architecture: 
        Tasker creates Tasks and places them in postgres table 'tasks'


