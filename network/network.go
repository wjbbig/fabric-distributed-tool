package network

import (
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"strconv"
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
	Peers      []string `yaml:"peers,omitempty"`
	Orderers   []string `yaml:"orderers,omitempty"`
	Chaincodes []string `yaml:"chaincodes,omitempty"`
}

type Chaincode struct {
	Path      string `yaml:"path,omitempty"`
	Version   string `yaml:"version,omitempty"`
	InitParam string `yaml:"init_param,omitempty"`
}

func GenerateNetworkConfig(fileDir, networkName, channelId, ccId, ccPath, ccVersion, ccInitParam string, peerUrls, ordererUrls []string) error {
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
	channel := &Channel{}
	nodes := make(map[string]*Node)
	for _, url := range peerUrls {
		node, err := NewNode(url, "peer")
		if err != nil {
			return err
		}
		nodes[node.hostname] = node
		channel.Peers = append(channel.Peers, node.hostname)
	}
	for _, url := range ordererUrls {
		node, err := NewNode(url, "orderer")
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
