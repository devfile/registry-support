module github.com/devfile/registry-support/index/server

go 1.14

require (
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/devfile/api/v2 v2.0.0-20220117162434-6e6e6a8bc14c
	github.com/devfile/library v1.2.1-0.20220308191614-f0f7e11b17de
	github.com/devfile/registry-support/index/generator v0.0.0-20220316161530-f06d84c42b54
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/go-version v1.4.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/client_golang v1.11.0
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/text v0.3.6
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
	k8s.io/apimachinery v0.21.3
)
