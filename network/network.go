package network

import (
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
)

const (
	defaultNetworkConfigName = "networkconfig.yaml"
	peerNode                 = "peer"
	ordererNode              = "orderer"
)

type NetworkConfig struct {
	Name       string                `yaml:"name,omitempty"`
	Channels   map[string]*Channel   `yaml:"channels,omitempty"`
	Nodes      map[string]*Node      `yaml:"nodes,omitempty"`
	Chaincodes map[string]*Chaincode `yaml:"chaincodes,omitempty"`
}

type Node struct {
	hostname string
	Name     string `yaml:"name,omitempty"`
	NodePort int    `json:"node_port,omitempty"`
	Type     string `yaml:"type,omitempty"`
	OrgId    string `yaml:"org_id,omitempty"`
	Domain   string `yaml:"domain,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Host     string `yaml:"host,omitempty"`
	SSHPort  int    `yaml:"ssh_port,omitempty"`
	Couch    bool   `yaml:"couch,omitempty"`
}

type Channel struct {
	Consensus  string   `yaml:"consensus,omitempty"`
	Peers      []string `yaml:"peers,omitempty"`
	Orderers   []string `yaml:"orderers,omitempty"`
	Chaincodes []string `yaml:"chaincodes,omitempty"`
}

type Chaincode struct {
	Path      string `yaml:"path,omitempty"`
	Version   string `yaml:"version,omitempty"`
	InitParam string `yaml:"init_param,omitempty"`
}

func (nc *NetworkConfig) GetPeerNodes() (peerNodes []*Node) {
	for _, node := range nc.Nodes {
		if node.Type == peerNode {
			peerNodes = append(peerNodes, node)
		}
	}
	return
}

func (nc *NetworkConfig) GetOrdererNodes() (ordererNodes []*Node) {
	for _, node := range nc.Nodes {
		if node.Type == ordererNode {
			ordererNodes = append(ordererNodes, node)
		}
	}
	return
}

func GenerateNetworkConfig(fileDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitParam string, peerUrls, ordererUrls []string) error {
	var network NetworkConfig

	network.Name = networkName
	chaincode := &Chaincode{
		Path:      ccPath,
		Version:   ccVersion,
		InitParam: ccInitParam,
	}
	network.Chaincodes = map[string]*Chaincode{
		ccId: chaincode,
	}
	channels := make(map[string]*Channel)
	channel := &Channel{Consensus: consensus}
	nodes := make(map[string]*Node)
	for _, url := range peerUrls {
		node, err := NewNode(url, peerNode)
		if err != nil {
			return err
		}
		nodes[node.hostname] = node
		channel.Peers = append(channel.Peers, node.hostname)
	}
	for _, url := range ordererUrls {
		node, err := NewNode(url, ordererNode)
		if err != nil {
			return err
		}
		nodes[node.hostname] = node
		channel.Orderers = append(channel.Orderers, node.hostname)
	}
	channel.Chaincodes = append(channel.Chaincodes, ccId)
	channels[channelId] = channel
	network.Channels = channels
	network.Nodes = nodes
	data, err := yaml.Marshal(network)
	if err != nil {
		return errors.Wrap(err, "failed to yaml marshal network config")
	}
	filePath := filepath.Join(fileDir, defaultNetworkConfigName)
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write networkconfig.yaml")
	}
	return nil
}

func NewNode(url string, nodeType string) (*Node, error) {
	hostname, nodePortStr, username, host, sshPortStr, password := utils.SplitUrlParam(url)
	nodePort, err := strconv.Atoi(nodePortStr)
	if err != nil {
		return nil, err
	}
	sshPort, err := strconv.Atoi(sshPortStr)
	if err != nil {
		return nil, err
	}
	name, orgId, domain := utils.SplitNameOrgDomain(hostname)
	return &Node{
		hostname: hostname,
		Name:     name,
		OrgId:    orgId,
		Domain:   domain,
		NodePort: nodePort,
		Username: username,
		Type:     nodeType,
		Password: password,
		Host:     host,
		SSHPort:  sshPort,
	}, nil
}
