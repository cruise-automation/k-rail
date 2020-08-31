module github.com/cruise-automation/k-rail

go 1.12

require (
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/mux v1.7.3
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/prometheus/common v0.13.0
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2 //v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab
	sigs.k8s.io/yaml v1.1.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
