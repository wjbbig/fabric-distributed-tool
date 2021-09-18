package network

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
)

const (
	defaultNetworkConfigName = "networkconfig.yaml"
	PeerNode                 = "peer"
	OrdererNode              = "orderer"
)

type NetworkConfig struct {
	Name       string                `yaml:"name,omitempty"`
	Version    string                `yaml:"version,omitempty"`
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
	Consensus  string              `yaml:"consensus,omitempty"`
	Peers      []string            `yaml:"peers,omitempty"`
	Orderers   []string            `yaml:"orderers,omitempty"`
	Chaincodes []*ChannelChaincode `yaml:"chaincodes,omitempty"`
}

type Chaincode struct {
	Path         string `yaml:"path,omitempty"`
	Version      string `yaml:"version,omitempty"`
	Policy       string `yaml:"policy,omitempty"`
	InitRequired bool   `yaml:"init_required,omitempty"`
	InitFunc     string `yaml:"init_func,omitempty"`
	InitParam    string `yaml:"init_param,omitempty"`
}

type ChannelChaincode struct {
	Name     string `yaml:"name,omitempty"`
	Sequence int64  `yaml:"sequence,omitempty"`
}

func (nc *NetworkConfig) GetPeerNodes() (peerNodes []*Node) {
	for _, node := range nc.Nodes {
		if node.Type == PeerNode {
			node.hostname = fmt.Sprintf("%s.%s", node.Name, node.Domain)
			peerNodes = append(peerNodes, node)
		}
	}
	return
}

func (nc *NetworkConfig) GetOrdererNodes() (ordererNodes []*Node) {
	for _, node := range nc.Nodes {
		if node.Type == OrdererNode {
			node.hostname = fmt.Sprintf("%s.%s", node.Name, node.Domain)
			ordererNodes = append(ordererNodes, node)
		}
	}
	return
}

func (nc *NetworkConfig) GetNodesByChannel(channelId string) (peerNodes []*Node, ordererNodes []*Node, err error) {
	channel, exists := nc.Channels[channelId]
	if !exists {
		err = errors.Errorf("%s not exists", channelId)
		return
	}

	for _, peerName := range channel.Peers {
		nc.Nodes[peerName].hostname = peerName
		peerNodes = append(peerNodes, nc.Nodes[peerName])
	}
	for _, ordererName := range channel.Orderers {
		nc.Nodes[ordererName].hostname = ordererName
		ordererNodes = append(ordererNodes, nc.Nodes[ordererName])
	}

	return
}

// ExtendChannelChaincode installs a new chaincode on channel
func (nc *NetworkConfig) ExtendChannelChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam string, initRequired bool) error {
	chaincode, exist := nc.Chaincodes[ccId]
	if exist {
		return errors.Errorf("chaincode %s exists", ccId)
	}
	chaincode = &Chaincode{
		Path:         ccPath,
		Version:      ccVersion,
		Policy:       ccPolicy,
		InitRequired: initRequired,
		InitFunc:     initFunc,
		InitParam:    initParam,
	}
	nc.Chaincodes[ccId] = chaincode
	channel, exist := nc.Channels[channelId]
	if !exist {
		return errors.Errorf("channel %s dose not exist", channelId)
	}
	channel.Chaincodes = append(channel.Chaincodes, &ChannelChaincode{
		Name:     ccId,
		Sequence: 1,
	})
	nc.Channels[channelId] = channel
	return writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) UpgradeChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam string, initRequired bool) error {
	chaincode, exist := nc.Chaincodes[ccId]
	if !exist {
		return errors.Errorf("chaincode %s does not exist", ccId)
	}
	if ccPath != "" {
		chaincode.Path = ccPath
	}
	if ccPolicy != "" {
		chaincode.Policy = ccPolicy
	}
	if initParam != "" {
		chaincode.InitParam = initParam
	}
	if initFunc != "" {
		chaincode.InitFunc = initFunc
	}
	// args must be defined
	chaincode.Version = ccVersion
	chaincode.InitRequired = initRequired

	channel, exist := nc.Channels[channelId]
	if !exist {
		return errors.Errorf("channel %s dose not exist", channelId)
	}
	for _, channelChaincode := range channel.Chaincodes {
		if channelChaincode.Name == ccId {
			channelChaincode.Sequence++
			break
		}
	}
	return writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) ExtendChannel(dataDir, channelId, consensus string, peers, orderers []string) error {
	_, exist := nc.Channels[channelId]
	if exist {
		return errors.Errorf("channel %s exists", channelId)
	}
	for _, peer := range peers {
		_, exist := nc.Nodes[peer]
		if !exist {
			return errors.Errorf("peer %s does not exist", peer)
		}
	}
	for _, orderer := range orderers {
		_, exist := nc.Nodes[orderer]
		if !exist {
			return errors.Errorf("orderer %s does not exist", orderer)
		}
	}
	channel := &Channel{
		Consensus: consensus,
		Peers:     peers,
		Orderers:  orderers,
	}
	nc.Channels[channelId] = channel
	return writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) ExtendNode(dataDir string, couchdb bool, peers, orderers []string) (peerNodes []*Node, ordererNodes []*Node, err error) {
	for _, peer := range peers {
		node, err := NewNode(peer, PeerNode, couchdb)
		if err != nil {
			return nil, nil, err
		}
		_, exist := nc.Nodes[node.hostname]
		if exist {
			return nil, nil, errors.Errorf("peer %s exists", node.hostname)
		}
		nc.Nodes[node.hostname] = node
		peerNodes = append(peerNodes, node)
	}
	for _, orderer := range orderers {
		node, err := NewNode(orderer, OrdererNode, false)
		if err != nil {
			return nil, nil, err
		}
		_, exist := nc.Nodes[node.hostname]
		if exist {
			return nil, nil, errors.Errorf("orderer %s exists", node.hostname)
		}
		nc.Nodes[node.hostname] = node
		ordererNodes = append(ordererNodes, node)
	}
	return peerNodes, ordererNodes, writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) GetNode(nodeName string) *Node {
	node, ok := nc.Nodes[nodeName]
	if !ok {
		return nil
	}

	node.hostname = fmt.Sprintf("%s.%s", node.Name, node.Domain)
	return node
}

func GenerateNetworkConfig(fileDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy string, ccInitRequired bool, sequence int64, couchdb bool, peerUrls, ordererUrls []string) (*NetworkConfig, error) {
	network := &NetworkConfig{}

	network.Name = networkName
	chaincode := &Chaincode{
		Path:         ccPath,
		Policy:       ccPolicy,
		Version:      ccVersion,
		InitRequired: ccInitRequired,
		InitParam:    ccInitParam,
		InitFunc:     ccInitFunc,
	}
	network.Chaincodes = map[string]*Chaincode{
		ccId: chaincode,
	}
	channels := make(map[string]*Channel)
	channel := &Channel{Consensus: consensus}
	nodes := make(map[string]*Node)
	for _, url := range peerUrls {
		node, err := NewNode(url, PeerNode, couchdb)
		if err != nil {
			return nil, err
		}
		nodes[node.hostname] = node
		channel.Peers = append(channel.Peers, node.hostname)
	}
	for _, url := range ordererUrls {
		node, err := NewNode(url, OrdererNode, couchdb)
		if err != nil {
			return nil, err
		}
		nodes[node.hostname] = node
		channel.Orderers = append(channel.Orderers, node.hostname)
	}
	channel.Chaincodes = append(channel.Chaincodes, &ChannelChaincode{
		Name:     ccId,
		Sequence: sequence,
	})
	channels[channelId] = channel
	network.Channels = channels
	network.Nodes = nodes
	if err := writeNetworkConfig(fileDir, network); err != nil {
		return nil, err
	}
	return network, nil
}

func writeNetworkConfig(dataDir string, nc *NetworkConfig) error {
	filePath := filepath.Join(dataDir, defaultNetworkConfigName)
	data, err := yaml.Marshal(nc)
	if err != nil {
		return errors.Wrapf(err, "yaml marshal failed")
	}
	if err := utils.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrapf(err, "write network config file failed")
	}
	return nil
}

func UnmarshalNetworkConfig(fileDir string) (*NetworkConfig, error) {
	filePath := filepath.Join(fileDir, defaultNetworkConfigName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading networkconfig file")
	}
	networkConfig := &NetworkConfig{}
	if err := yaml.Unmarshal(data, networkConfig); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling networkConfig")
	}
	return networkConfig, nil
}

func NewNode(url string, nodeType string, couchdb bool) (*Node, error) {
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
	if nodeType == OrdererNode {
		couchdb = false
	}
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
		Couch:    couchdb,
	}, nil
}

func (n *Node) GetHostname() string {
	return n.hostname
}
