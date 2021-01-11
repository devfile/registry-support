module github.com/devfile/registry-support/index/server

go 1.14

require (
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/devfile/registry-support/index/generator v0.0.0-20201211145454-bc904934d987
	github.com/gin-gonic/gin v1.6.3
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/client_golang v1.9.0
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/apimachinery v0.19.4
)
