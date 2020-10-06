# oci-devfile-registry

Simple proof of concept devfile registry using an OCI-based registry for storage on Kubernetes.

## Deploy

### OpenShift

1) `cd deploy/openshift`
2) Set the value of `$HOST` in `route.yaml` to your routing suffix (such as apps.devcluster.example.com)
3) Run `oc apply -f registry.yaml`
4) Run `oc apply -f route.yaml`