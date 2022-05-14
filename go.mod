module github.com/wjbbig/fabric-distributed-tool

go 1.16

require (
	github.com/docker/docker v20.10.7+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/hyperledger/fabric v2.1.1+incompatible
	github.com/hyperledger/fabric-config v0.1.0
	github.com/hyperledger/fabric-protos-go v0.0.0-20210528200356-82833ecdac31
	github.com/hyperledger/fabric-sdk-go v1.0.1-0.20210603143513-14047c6d88f0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.10.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/containerd/containerd v1.5.2 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20210603140002-2670f91851c8 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
)

replace github.com/mitchellh/mapstructure v1.4.1 => github.com/mitchellh/mapstructure v1.3.3
