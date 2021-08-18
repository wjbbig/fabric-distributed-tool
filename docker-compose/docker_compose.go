package docker_compose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
)

var logger = mylogger.NewLogger()

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
	DependsOn     []string `yaml:"depends_on,omitempty"`
	Networks      []string `yaml:"networks,omitempty"`
	ExtraHosts    []string `yaml:"extra_hosts,omitempty"`
}

type ExternalNetwork struct {
	External bool `yaml:"external,omitempty"`
}

// GenerateOrdererDockerComposeFile 生成启动orderer的docker-compose文件
func GenerateOrdererDockerComposeFile(filePath string, orderer *network.Node, otherUrls []string) error {
	var dockerCompose DockerCompose
	logger.Infof("begin to generate docker_compose file, url=%s", orderer.GetHostname())

	ordererService := Service{
		ContainerName: orderer.GetHostname(),
		Image:         "hyperledger/fabric-orderer",
		Environment: []string{
			"FABRIC_LOGGING_SPEC=INFO",
			"ORDERER_GENERAL_LISTENADDRESS=0.0.0.0",
			"ORDERER_GENERAL_GENESISMETHOD=file",
			"ORDERER_GENERAL_GENESISFILE=/var/hyperledger/orderer/orderer.genesis.block",
			fmt.Sprintf("ORDERER_GENERAL_LOCALMSPID=%s", orderer.OrgId),
			"ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp",
			"ORDERER_GENERAL_TLS_ENABLED=true",
			"ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key",
			"ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt",
			"ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]",
			"ORDERER_KAFKA_TOPIC_REPLICATIONFACTOR=1",
			"ORDERER_KAFKA_VERBOSE=true",
			"ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=/var/hyperledger/orderer/tls/server.crt",
			"ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=/var/hyperledger/orderer/tls/server.key",
			"ORDERER_GENERAL_CLUSTER_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]",
		},
		WorkingDir: "/opt/gopath/src/github.com/hyperledger/fabric",
		Command:    "orderer",
		Volumes: []string{
			fmt.Sprintf("%s/channel-artifacts/genesis.block:/var/hyperledger/orderer/orderer.genesis.block", filePath),
			fmt.Sprintf("%s/crypto-config/ordererOrganizations/%s/orderers/%s/msp:/var/hyperledger/orderer/msp",
				filePath, orderer.Domain, orderer.GetHostname()),
			fmt.Sprintf("%s/crypto-config/ordererOrganizations/%s/orderers/%s/tls/:/var/hyperledger/orderer/tls",
				filePath, orderer.Domain, orderer.GetHostname()),
			// fmt.Sprintf("%s:/var/hyperledger/production/orderer", ordererURLArgs[0]),
		},
		Ports:      []string{fmt.Sprintf("%[1]d:7050", orderer.NodePort)},
		Networks:   []string{defaultNetworkName},
		ExtraHosts: otherUrls,
	}
	dockerCompose.Services = map[string]Service{
		orderer.GetHostname(): ordererService,
	}
	dockerCompose.Version = `2`
	dockerCompose.Networks = map[string]ExternalNetwork{
		defaultNetworkName: {},
	}
	_, err := os.Stat(filePath)
	if err != nil {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			return err
		}
	}
	filePath = filepath.Join(filePath, fmt.Sprintf("%s%s.yaml", defaultDockerComposeFile,
		strings.ReplaceAll(orderer.GetHostname(), ".", "-")))
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrapf(err, "failed to write peer docker_compose file, url=%s", orderer.GetHostname())
	}
	logger.Infof("finish generating peer docker_compose file, url=%s", orderer.GetHostname())
	return nil
}

// GeneratePeerDockerComposeFile 生产peer的docker-compose启动文件
func GeneratePeerDockerComposeFile(filePath string, peer *network.Node, gossipBootstrapPeerUrl string, otherUrls []string, couchdb bool) error {
	var dockerCompose DockerCompose
	logger.Infof("begin to generate peer docker_compose file, url=%s", peer.GetHostname())
	networkPrefix := path.Base(filePath)
	peerService := Service{
		ContainerName: peer.GetHostname(),
		Image:         "hyperledger/fabric-peer",
		WorkingDir:    "/opt/gopath/src/github.com/hyperledger/fabric/peer",
		Environment: []string{
			"CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock",
			fmt.Sprintf("CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=%s_fabric_network", networkPrefix),
			"FABRIC_LOGGING_SPEC=INFO",
			"CORE_PEER_TLS_ENABLED=true",
			"CORE_PEER_GOSSIP_USELEADERELECTION=true",
			"CORE_PEER_GOSSIP_ORGLEADER=false",
			"CORE_PEER_PROFILE_ENABLED=true",
			"CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt",
			"CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key",
			"CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt",
			fmt.Sprintf("CORE_PEER_ID=%s", peer.GetHostname()),
			fmt.Sprintf("CORE_PEER_ADDRESS=%s:%d", peer.GetHostname(), peer.NodePort),
			fmt.Sprintf("CORE_PEER_LISTENADDRESS=0.0.0.0:%d", peer.NodePort),
			fmt.Sprintf("CORE_PEER_CHAINCODEADDRESS=%s:%d", peer.GetHostname(), peer.NodePort+1),
			fmt.Sprintf("CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:%d", peer.NodePort+1),
			fmt.Sprintf("CORE_PEER_GOSSIP_BOOTSTRAP=%s", gossipBootstrapPeerUrl),
			fmt.Sprintf("CORE_PEER_GOSSIP_EXTERNALENDPOINT=%s:%d", peer.GetHostname(), peer.NodePort),
			fmt.Sprintf("CORE_PEER_LOCALMSPID=%s", peer.OrgId),
		},
		Command: "peer node start",
		Volumes: []string{
			"/var/run/:/host/var/run/",
			fmt.Sprintf("%s/crypto-config/peerOrganizations/%s/peers/%s/msp:/etc/hyperledger/fabric/msp", filePath, peer.Domain, peer.GetHostname()),
			fmt.Sprintf("%s/crypto-config/peerOrganizations/%s/peers/%s/tls:/etc/hyperledger/fabric/tls", filePath, peer.Domain, peer.GetHostname()),
			// fmt.Sprintf("%s:/var/hyperledger/production", peerUrlArgs[0]),
		},
		Ports:      []string{fmt.Sprintf("%[1]d:%[1]d", peer.NodePort)},
		Networks:   []string{defaultNetworkName},
		ExtraHosts: otherUrls,
	}
	dockerCompose.Version = `2`
	dockerCompose.Networks = map[string]ExternalNetwork{
		defaultNetworkName: {},
	}
	// create the directory first
	_, err := os.Stat(filePath)
	if err != nil {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			return err
		}
	}
	// generate docker compose file if using couchdb
	if couchdb {
		couchdbServiceName, err := GenerateCouchDB(filePath, peer.GetHostname())
		if err != nil {
			return err
		}
		peerService.Environment = append(peerService.Environment, "CORE_LEDGER_STATE_STATEDATABASE=CouchDB")
		peerService.Environment = append(peerService.Environment, fmt.Sprintf("CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=%s:5984", couchdbServiceName))
		peerService.Environment = append(peerService.Environment, "CORE_LEDGER_STATE_COUCHDBCONFIG_USERNAME=")
		peerService.Environment = append(peerService.Environment, "CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD=")
	}
	dockerCompose.Services = map[string]Service{
		peer.GetHostname(): peerService,
	}
	filePath = filepath.Join(filePath, fmt.Sprintf("%s%s.yaml", defaultDockerComposeFile,
		strings.ReplaceAll(peer.GetHostname(), ".", "-")))
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrapf(err, "failed to write peer docker_compose file, url=%s", peer.GetHostname())
	}
	logger.Infof("finish generating peer docker_compose file, url=%s", peer.GetHostname())
	return nil
}

// GenerateCLI generates the docker compose file for cli container.
// if the fabric version is 2.x, this will not be generated.
// if the endpoint has only one orderer node, this file will not be generated too, thus orderer does not need it.
// if the endpoint has two or more peer nodes, then only generates one cli file, cause you can connect to other peers
// by changing env params
func GenerateCLI(filePath string) error {
	return nil
}

func GenerateCA(filePath string, orgId string, domain string, port string) error {
	var dockerCompose DockerCompose
	dockerCompose.Version = "2"
	caService := Service{
		ContainerName: fmt.Sprintf("ca_%s", orgId),
		Image:         "hyperledger/fabric-ca",
		Environment: []string{
			"FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server",
			fmt.Sprintf("FABRIC_CA_SERVER_CA_NAME=ca-%s", orgId),
			"FABRIC_CA_SERVER_TLS_ENABLED=true",
			fmt.Sprintf("FABRIC_CA_SERVER_TLS_CERTFILE=/etc/hyperledger/fabric-ca-server-config/ca.%s-cert.pem", domain),
			"FABRIC_CA_SERVER_TLS_KEYFILE=/etc/hyperledger/fabric-ca-server-config/priv_sk",
			fmt.Sprintf("FABRIC_CA_SERVER_PORT=%s", port),
		},
		Command: `sh -c 'fabric-ca-server start --ca.certfile /etc/hyperledger/fabric-ca-server-config/ca.org1.example.com-cert.pem --ca.keyfile /etc/hyperledger/fabric-ca-server-config/priv_sk -b admin:adminpw -d'`,
		Volumes: []string{
			fmt.Sprintf("%s/crypto-config/peerOrganizations/%s/ca/:/etc/hyperledger/fabric-ca-server-config",
				filePath, domain),
		},
		Ports:    []string{fmt.Sprintf("%[1]s:%[1]s", port)},
		Networks: []string{defaultNetworkName},
	}
	dockerCompose.Services = map[string]Service{
		fmt.Sprintf("ca_%s", orgId): caService,
	}
	dockerCompose.Networks = map[string]ExternalNetwork{
		defaultNetworkName: {},
	}
	caFileName := fmt.Sprintf("%s%s-ca.yaml", defaultDockerComposeFile, domain)
	filePath = filepath.Join(filePath, caFileName)
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		return errors.Wrapf(err, "failed to generate docker-compose file for %s ca", orgId)
	}
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	logger.Infof("finish generating peer docker_compose_ca file, org=%s", orgId)
	return nil
}

func GenerateCouchDB(filePath string, peerUrl string) (string, error) {
	logger.Infof("begin to generate couchdb docker compose file for %s", peerUrl)
	var dockerCompose DockerCompose
	dockerCompose.Version = "2"
	dockerCompose.Networks = map[string]ExternalNetwork{
		defaultNetworkName: {},
	}
	//name, org, _ := utils.SplitNameOrgDomain(peerUrl)
	port := utils.GetRandomPort()
	serviceName := fmt.Sprintf("couchdb%d", port)
	couchDBService := Service{
		ContainerName: serviceName,
		Image:         "hyperledger/fabric-couchdb",
		Environment: []string{
			"COUCHDB_USER=",
			"COUCHDB_PASSWORD=",
		},
		Ports:    []string{fmt.Sprintf("%d:%d", port, 5984)},
		Networks: []string{defaultNetworkName},
	}
	dockerCompose.Services = map[string]Service{
		serviceName: couchDBService,
	}

	filePath = filepath.Join(filePath, fmt.Sprintf("%s%s-couchdb.yaml", defaultDockerComposeFile, strings.ReplaceAll(peerUrl, ".", "-")))
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		return "", errors.Wrap(err, "failed to yaml marshal couchdb docker compose file")
	}
	if err = ioutil.WriteFile(filePath, data, 0755); err != nil {
		return "", errors.Wrap(err, "failed to write file")
	}
	logger.Infof("finish generating couchdb docker_compose file for %s", peerUrl)
	return serviceName, nil
}
