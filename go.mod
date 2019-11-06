module github.com/cruise-automation/k-rail

go 1.12

require (
	github.com/gobwas/glob v0.2.3
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191105190043-25240d7d6d90
	k8s.io/apimachinery v0.0.0-20191105185716-00d39968b57e
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
