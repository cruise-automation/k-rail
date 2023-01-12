module github.com/cruise-automation/k-rail/v3

go 1.16

require (
	github.com/gertd/go-pluralize v0.2.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-plugin v1.4.0
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.6.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	google.golang.org/grpc v1.35.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	istio.io/api v0.0.0-20210817134638-1b724d2c575d
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2 //v11.0.0+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/yaml v1.2.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
