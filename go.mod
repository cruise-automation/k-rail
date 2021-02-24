module github.com/cruise-automation/k-rail

go 1.12

require (
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/mux v1.7.3
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/sirupsen/logrus v1.6.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2 //v11.0.0+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/yaml v1.2.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
