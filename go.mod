module github.com/infobloxopen/themis

go 1.15

require (
	github.com/allegro/bigcache/v2 v2.2.4
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/infobloxopen/go-trees v0.0.0-20190313150506-2af4e13f9062
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.1.0 // indirect
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
)

replace (
	// Required by coredns/coredns https://github.com/coredns/coredns/blob/v1.7.1/go.mod#L34
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/api => k8s.io/api v0.19.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3
	k8s.io/client-go => k8s.io/client-go v0.19.3
)
