module github.com/devfile/registry-support/index/server

go 1.14

require (
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/devfile/api/v2 v2.0.0-20211021164004-dabee4e633ed
	github.com/devfile/registry-support/index/generator v0.0.0-20220222194908-7a90a4214f3e
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/client_golang v1.11.0
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/text v0.3.6
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
	k8s.io/apimachinery v0.21.3
)
