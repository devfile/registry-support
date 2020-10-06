# oci-devfile-registry

Simple proof of concept devfile registry using an OCI-based registry for storage on Kubernetes.

## Architecture
![Test Image 1](https://user-images.githubusercontent.com/606959/88183913-5e417280-cc32-11ea-9509-651bb44d9280.png)

## Deploy

### OpenShift

1) `cd deploy/openshift`
2) Set the value of `$HOST` in `route.yaml` to your routing suffix (such as apps.devcluster.example.com)
3) Run `oc apply -f registry.yaml`
4) Run `oc apply -f route.yaml`
