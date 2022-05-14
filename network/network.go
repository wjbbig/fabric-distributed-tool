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
	DefaultNetworkConfigName = "networkconfig.yaml"
	PeerNode                 = "peer"
	OrdererNode              = "orderer"
	CANode                   = "ca"
	defaultV14ImageTag       = "1.4.4"
	defaultV20ImageTag       = "2.4.3"
	defaultCAImageTag        = "1.4.9"
	defaultCouchDBImageTag   = "0.4.22"
	defaultCAEnrollId        = "admin"
	defaultCAEnrollSecret    = "adminpw"
)

type NetworkConfig struct {
	Consensus     string                `yaml:"consensus,omitempty"`
	Name          string                `yaml:"name,omitempty"`
	Version       string                `yaml:"version,omitempty"`
	NodeImageTag  string                `yaml:"node_image_tag,omitempty"`
	CAImageTag    string                `yaml:"ca_image_tag,omitempty"`
	CouchImageTag string                `yaml:"couch_image_tag,omitempty"`
	Channels      map[string]*Channel   `yaml:"channels,omitempty"`
	Nodes         map[string]*Node      `yaml:"nodes,omitempty"`
	Cas           map[string]*CA        `yaml:"cas,omitempty"`
	Chaincodes    map[string]*Chaincode `yaml:"chaincodes,omitempty"`
}

type Node struct {
	hostname    string
	Name        string `yaml:"name,omitempty"`
	NodePort    int    `json:"node_port,omitempty"`
	Type        string `yaml:"type,omitempty"`
	OrgId       string `yaml:"org_id,omitempty"`
	Domain      string `yaml:"domain,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
	Host        string `yaml:"host,omitempty"`
	SSHPort     int    `yaml:"ssh_port,omitempty"`
	Couch       bool   `yaml:"couch,omitempty"`
	Dest        string `yaml:"dest,omitempty"`
	Transferred bool   `yaml:"transferred,omitempty"` // check if the file of node has been transferred to dest dir
	Start       bool   `yaml:"start,omitempty"`
}

type CA struct {
	Name         string `json:"name,omitempty"`
	EnrollId     string `json:"enroll_id,omitempty"`
	EnrollSecret string `json:"enroll_secret,omitempty"`
}

type Channel struct {
	Peers      []string            `yaml:"peers,omitempty"`
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

func (nc *NetworkConfig) GetOrgDomain(orgId string) string {
	for _, node := range nc.Nodes {
		if node.OrgId == orgId && node.Type == PeerNode {
			return node.Domain
		}
	}

	return ""
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

func (nc *NetworkConfig) GetCANodes() (caNodes []*Node) {
	for _, node := range nc.Nodes {
		if node.Type == CANode {
			node.hostname = node.Name
			caNodes = append(caNodes, node)
		}
	}
	return
}

func (nc *NetworkConfig) GetCAByOrgId(orgName string) (*Node, *CA) {
	caName := fmt.Sprintf("ca_%s", orgName)
	return nc.Nodes[caName], nc.Cas[caName]
}

func (nc *NetworkConfig) CreateCANode(dataDir, orgId, enrollId, enrollSecret string, port int) error {
	caName := fmt.Sprintf("ca_%s", orgId)
	if _, exists := nc.Nodes[caName]; exists {
		return errors.Errorf("ca node %s exist", caName)
	}
	nodes := nc.GetPeerNodes()
	var node Node
	for _, n := range nodes {
		if n.OrgId == orgId {
			node = *n
		}
	}
	node.NodePort = port
	node.Couch = false
	node.Type = CANode
	node.Name = caName
	nc.Nodes[caName] = &node
	if enrollId == "" {
		enrollId = defaultCAEnrollId
	}
	if enrollSecret == "" {
		enrollSecret = defaultCAEnrollSecret
	}
	ca := &CA{caName, enrollId, enrollSecret}
	if nc.Cas == nil {
		nc.Cas = make(map[string]*CA)
	}
	nc.Cas[caName] = ca
	return writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) GetNodesByChannel(channelId string) (peerNodes []*Node, ordererNodes []*Node, err error) {
	channel, exists := nc.Channels[channelId]
	if !exists {
		err = errors.Errorf("channel %s does not exists", channelId)
		return
	}

	for _, peerName := range channel.Peers {
		nc.Nodes[peerName].hostname = peerName
		peerNodes = append(peerNodes, nc.Nodes[peerName])
	}

	ordererNodes = nc.GetOrdererNodes()

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

func (nc *NetworkConfig) ExtendChannel(dataDir, channelId string, peers []string) error {
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

	channel := &Channel{
		Peers: peers,
	}
	nc.Channels[channelId] = channel
	return writeNetworkConfig(dataDir, nc)
}

func (nc *NetworkConfig) ExtendNode(dataDir string, couchdb bool, peers, orderers []string) (peerNodes []*Node, ordererNodes []*Node, err error) {
	for _, peer := range peers {
		node, err := NewNode(peer, PeerNode, couchdb, dataDir)
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
		node, err := NewNode(orderer, OrdererNode, false, dataDir)
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

func (nc *NetworkConfig) GetImageTags() []string {
	return []string{nc.NodeImageTag, nc.CAImageTag, nc.CouchImageTag}
}

func (nc *NetworkConfig) GetNode(nodeName string) *Node {
	node, ok := nc.Nodes[nodeName]
	if !ok {
		return nil
	}

	node.hostname = fmt.Sprintf("%s.%s", node.Name, node.Domain)
	return node
}

func GenerateNetworkConfig(fileDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy string, ccInitRequired bool, sequence int64, couchdb bool, peerUrls, ordererUrls []string, networkVersion string) (*NetworkConfig, error) {
	if err := validateBasic(fileDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy, sequence, peerUrls, ordererUrls, networkVersion); err != nil {
		return nil, err
	}
	network := &NetworkConfig{
		Name:          networkName,
		Consensus:     consensus,
		Version:       networkVersion,
		CouchImageTag: defaultCouchDBImageTag,
		CAImageTag:    defaultCAImageTag,
	}
	if networkVersion == "v1.4" {
		network.NodeImageTag = defaultV14ImageTag
	} else {
		network.NodeImageTag = defaultV20ImageTag
	}
	if ccId != "" {
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
	}
	channels := make(map[string]*Channel)
	channel := &Channel{}
	nodes := make(map[string]*Node)
	for _, url := range peerUrls {
		node, err := NewNode(url, PeerNode, couchdb, fileDir)
		if err != nil {
			return nil, err
		}
		nodes[node.hostname] = node
		channel.Peers = append(channel.Peers, node.hostname)
	}
	for _, url := range ordererUrls {
		node, err := NewNode(url, OrdererNode, false, fileDir)
		if err != nil {
			return nil, err
		}
		nodes[node.hostname] = node
	}
	if ccId != "" {
		channel.Chaincodes = append(channel.Chaincodes, &ChannelChaincode{
			Name:     ccId,
			Sequence: sequence,
		})
	}
	channels[channelId] = channel
	network.Channels = channels
	network.Nodes = nodes
	if err := writeNetworkConfig(fileDir, network); err != nil {
		return nil, err
	}
	return network, nil
}

func validateBasic(fileDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy string, sequence int64, peerUrls, ordererUrls []string, networkVersion string) error {
	if fileDir == "" {

	}
	return nil
}

func writeNetworkConfig(dataDir string, nc *NetworkConfig) error {
	filePath := filepath.Join(dataDir, DefaultNetworkConfigName)
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
	filePath := filepath.Join(fileDir, DefaultNetworkConfigName)
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

func NewNode(url string, nodeType string, couchdb bool, dest string) (*Node, error) {
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
		Couch:    couchdb,
		Dest:     dest,
	}, nil
}

func (n *Node) GetHostname() string {
	return n.hostname
}
