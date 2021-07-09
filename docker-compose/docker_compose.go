package docker_compose

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const defaultDockerComposeFile = "docker-compose-"
const defaultNetworkName = "fabric_network"

type DockerCompose struct {
	Version  string                     `yaml:"version"`
	Networks map[string]ExternalNetwork `yaml:"networks,omitempty"`
	Services map[string]Service         `yaml:"services"`
}

type Service struct {
	ContainerName string   `yaml:"container_name,omitempty"`
	Image         string   `yaml:"image,omitempty"`
	Environment   []string `yaml:"environment,omitempty"`
	WorkingDir    string   `yaml:"working_dir,omitempty"`
	Command       string   `yaml:"command,omitempty"`
	Volumes       []string `yaml:"volumes,omitempty"`
	Ports         []string `yaml:"ports,omitempty"`
	dependsOn     []string `yaml:"depends_on,omitempty"`
	Networks      []string `yaml:"networks,omitempty"`
	ExtraHosts    []string `yaml:"extra_hosts,omitempty"`
}

type ExternalNetwork struct {
	External bool `yaml:"external,omitempty"`
}

// detectImageNameAndTag 找到包含某个关键字的docker image的tag
func detectImageNameAndTag(keyword string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", errors.Wrapf(err, "detect image of %s failed", keyword)
	}
	defer cli.Close()
	options := types.ImageListOptions{
		All:     false,
		Filters: filters.Args{},
	}
	imageList, err := cli.ImageList(context.Background(), options)
	if err != nil {
		return "", errors.Wrapf(err, "detect image of %s failed", keyword)
	}

	for _, summary := range imageList {
		if summary.RepoTags == nil {
			continue
		}
		if strings.Contains(summary.RepoTags[0], keyword) {
			return summary.RepoTags[0], nil
		}
	}

	return "", nil
}

func GenerateOrdererDockerComposeFile(filePath string, ordererUrl string, otherUrls []string) error {
	return nil
}

// GeneratePeerDockerComposeFile 生产peer的docker-compose启动文件
func GeneratePeerDockerComposeFile(filePath string, peerUrl string, anchorPeerUrl string, otherUrls []string) error {
	var dockerCompose DockerCompose
	imageName, err := detectImageNameAndTag("peer")
	if err != nil {
		return err
	}
	peerUrlArgs := strings.Split(peerUrl, ":")
	_, orgName, domain := util.SplitNameOrgDomain(peerUrlArgs[0])
	peerService := Service{
		ContainerName: peerUrlArgs[0],
		Image:         imageName,
		WorkingDir:    "/opt/gopath/src/github.com/hyperledger/fabric/peer",
		Environment: []string{
			"CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock",
			"CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=fabric_network",
			"FABRIC_LOGGING_SPEC=INFO",
			"CORE_PEER_TLS_ENABLED=true",
			"CORE_PEER_GOSSIP_USELEADERELECTION=true",
			"CORE_PEER_GOSSIP_ORGLEADER=false",
			"CORE_PEER_PROFILE_ENABLED=true",
			"CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt",
			"CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key",
			"CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt",
			fmt.Sprintf("CORE_PEER_ID=%s", peerUrlArgs[0]),
			fmt.Sprintf("CORE_PEER_ADDRESS=%s", peerUrl),
			fmt.Sprintf("CORE_PEER_LISTENADDRESS=0.0.0.0:%s", peerUrlArgs[1]),
			fmt.Sprintf("CORE_PEER_CHAINCODEADDRESS=%s:7052", peerUrlArgs[0]),
			"CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7052",
			fmt.Sprintf("CORE_PEER_GOSSIP_BOOTSTRAP=%s", anchorPeerUrl),
			fmt.Sprintf("CORE_PEER_GOSSIP_EXTERNALENDPOINT=%s", peerUrl),
			fmt.Sprintf("CORE_PEER_LOCALMSPID=%sMSP", orgName),
		},
		Command: "peer node start",
		Volumes: []string{
			"/var/run/:/host/var/run/",
			fmt.Sprintf("%s/crypto-config/peerOrganizations/%s/peers/%s/msp:/etc/hyperledger/fabric/msp", filePath, domain, peerUrlArgs[0]),
			fmt.Sprintf("%s/crypto-config/peerOrganizations/%s/peers/%s/tls:/etc/hyperledger/fabric/tls", filePath, domain, peerUrlArgs[0]),
			fmt.Sprintf("%s:/var/hyperledger/production", peerUrlArgs[0]),
		},
		Ports:      []string{fmt.Sprintf("%s:%s", peerUrlArgs[1], peerUrlArgs[1])},
		Networks:   []string{defaultNetworkName},
		ExtraHosts: nil,
	}
	dockerCompose.Services = map[string]Service{
		peerUrlArgs[0]: peerService,
	}
	dockerCompose.Version = `2`
	dockerCompose.Networks = map[string]ExternalNetwork{
		defaultNetworkName: {},
	}
	// 先创建文件夹
	_, err = os.Stat(filePath)
	if err != err {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			return err
		}
	}
	filePath = filepath.Join(filePath, fmt.Sprintf("%s%s.yaml", defaultDockerComposeFile,
		strings.ReplaceAll(peerUrlArgs[0], ".", "-")))
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0755)
}
