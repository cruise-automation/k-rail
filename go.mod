module github.com/cruise-automation/k-rail

go 1.12

require (
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.0.0-20200625001655-4c5254603344 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.18.10
	k8s.io/apimachinery v0.18.10
	k8s.io/client-go v0.18.10 //v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/yaml v1.2.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
