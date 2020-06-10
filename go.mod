module github.com/eclipse/codewind-installer

go 1.12

require (
	cloud.google.com/go v0.49.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest/autorest v0.10.2 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Microsoft/go-winio v0.4.13 // indirect
	github.com/containerd/containerd v1.3.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20191007211215-3e077fc8667a+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/google/go-github/v32 v32.0.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gophercloud/gophercloud v0.7.0 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/openshift/api v0.0.0-20191025141232-e7fa4b871a25
	github.com/openshift/client-go v0.0.0-20191022152013-2823239d2298
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.21.0
	github.com/zalando/go-keyring v0.0.0-20190913082157-62750a1ff80d
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20191023065245-6d3f0bb11be5 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.7
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.0.0-20191023225726-842530cfd124
	k8s.io/apimachinery v0.0.0-20191023225540-31cb258e7ad9
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c // indirect
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
)

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20191007211215-3e077fc8667a+incompatible
