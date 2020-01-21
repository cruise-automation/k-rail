module github.com/cruise-automation/k-rail

go 1.12

require (
	github.com/gobwas/glob v0.2.3
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
