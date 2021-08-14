package fabricconfig

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	fconfigtx "github.com/hyperledger/fabric-config/configtx"
	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigtxFileName   = "configtx.yaml"
	defaultConsortiumName     = "FabricConsortiums"
	defaultGenesisName        = "FabricGenesis"
	ordererType_SOLO          = "solo"
	ordererType_ETCDRAFT      = "etcdraft"
	defaultChannelProfileName = "FabricChannel"
	defaultGenesisChannel     = "fabric-genesis-channel"
)

type Configtx struct {
	Profiles map[string]ConfigtxProfile `yaml:"Profiles,omitempty"`
}

type ConfigtxProfile struct {
	Consortium   string                        `yaml:"Consortium,omitempty"`
	Orderer      ConfigtxOrderer               `yaml:"Orderer,omitempty"`
	Application  ConfigtxApplication           `yaml:"Application,omitempty"`
	Capabilities map[string]bool               `yaml:"Capabilities,omitempty"`
	Consortiums  map[string]ConfigtxConsortium `yaml:"Consortiums,omitempty"`
	Policies     map[string]ConfigtxPolicy     `yaml:"Policies,omitempty"`
}

type ConfigtxConsortium struct {
	Organizations []ConfigtxOrganization `yaml:"Organizations,omitempty"`
}

type ConfigtxOrderer struct {
	OrdererType   string                    `yaml:"OrdererType,omitempty"`
	Addresses     []string                  `yaml:"Addresses,omitempty"`
	BatchTimeout  string                    `yaml:"BatchTimeout,omitempty"`
	BatchSize     ConfigtxBatchSize         `yaml:"BatchSize,omitempty"`
	Kafka         ConfigtxKafka             `yaml:"Kafka,omitempty"`
	EtcdRaft      ConfigtxEtcdRaft          `yaml:"EtcdRaft,omitempty"`
	Organizations []ConfigtxOrganization    `yaml:"Organizations"`
	Policies      map[string]ConfigtxPolicy `yaml:"Policies,omitempty"`
	Capabilities  map[string]bool           `yaml:"Capabilities,omitempty"`
}

type ConfigtxEtcdRaft struct {
	Consenters []ConfigtxConsenter `yaml:"Consenters,omitempty"`
}

type ConfigtxConsenter struct {
	Host          string `yaml:"Host,omitempty"`
	Port          uint32 `yaml:"Port,omitempty"`
	ClientTLSCert string `yaml:"ClientTLSCert,omitempty"`
	ServerTLSCert string `yaml:"ServerTLSCert,omitempty"`
}

type ConfigtxBatchSize struct {
	MaxMessageCount   uint32 `yaml:"MaxMessageCount,omitempty"`
	AbsoluteMaxBytes  string `yaml:"AbsoluteMaxBytes,omitempty"`
	PreferredMaxBytes string `yaml:"PreferredMaxBytes,omitempty"`
}

type ConfigtxKafka struct {
	Brokers []string `yaml:"Brokers,omitempty"`
}

type ConfigtxPolicy struct {
	Type string `yaml:"Type,omitempty"`
	Rule string `yaml:"Rule,omitempty"`
}

type ConfigtxOrganization struct {
	Name        string                    `yaml:"Name,omitempty"`
	ID          string                    `yaml:"ID,omitempty"`
	MSPDir      string                    `yaml:"MSPDir,omitempty"`
	Policies    map[string]ConfigtxPolicy `yaml:"Policies,omitempty"`
	AnchorPeers []ConfigtxAnchorPeer      `yaml:"AnchorPeers,omitempty"`
}

type ConfigtxAnchorPeer struct {
	Host string `yaml:"Host,omitempty"`
	Port uint32 `yaml:"Port,omitempty"`
}

type ConfigtxApplication struct {
	Organizations []ConfigtxOrganization    `yaml:"Organizations,omitempty"`
	Policies      map[string]ConfigtxPolicy `yaml:"Policies,omitempty"`
	Capabilities  map[string]bool           `yaml:"Capabilities,omitempty"`
}

// orderPeerOrdererByOrg 将相同组织的节点整理在一起
func orderPeerOrdererByOrg(nodes []*network.Node) map[string][]*network.Node {
	orderedUrl := make(map[string][]*network.Node)
	for _, node := range nodes {
		orderedUrl[node.OrgId] = append(orderedUrl[node.OrgId], node)
	}
	return orderedUrl
}

func GenerateConfigtxFile(filePath string, ordererType string, orderers, peers []*network.Node) error {
	logger.Infof("begin to generate configtx.yaml, consensus type=%s", ordererType)
	defer logger.Info("finish generating configtx.yaml")
	var configtx Configtx
	var consenters []ConfigtxConsenter
	var ordererOrganizations []ConfigtxOrganization
	ordererOrgsPath := filepath.Join(filePath, "crypto-config", "ordererOrganizations")
	ordererMap := orderPeerOrdererByOrg(orderers)
	var ordererAddresses []string
	for _, ordererNodes := range ordererMap {
		for _, node := range ordererNodes {
			serverCertPath := filepath.Join(ordererOrgsPath, node.Domain, "orderers", node.GetHostname(), "tls/server.crt")
			consenter := ConfigtxConsenter{
				Host:          node.GetHostname(),
				Port:          uint32(node.NodePort),
				ClientTLSCert: serverCertPath,
				ServerTLSCert: serverCertPath,
			}
			consenters = append(consenters, consenter)
			ordererAddresses = append(ordererAddresses, fmt.Sprintf("%s:%d", node.GetHostname(), node.NodePort))
		}
		ordererNode := ordererNodes[0]

		ordererOrganization := ConfigtxOrganization{
			Name:   ordererNode.OrgId,
			ID:     ordererNode.OrgId,
			MSPDir: filepath.Join(ordererOrgsPath, ordererNode.Domain, "msp"),
			Policies: map[string]ConfigtxPolicy{
				fconfigtx.ReadersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin','%[1]s.orderer','%[1]s.client')", ordererNode.OrgId),
				},
				fconfigtx.WritersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.orderer', '%[1]s.client')", ordererNode.OrgId),
				},
				fconfigtx.AdminsPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%s.admin')", ordererNode.OrgId),
				},
			},
		}
		ordererOrganizations = append(ordererOrganizations, ordererOrganization)
	}

	orderer := ConfigtxOrderer{
		OrdererType:  ordererType,
		Addresses:    ordererAddresses,
		BatchTimeout: "2s",
		BatchSize: ConfigtxBatchSize{
			MaxMessageCount:   10,
			AbsoluteMaxBytes:  "99 MB",
			PreferredMaxBytes: "512 KB",
		},
		EtcdRaft:      ConfigtxEtcdRaft{Consenters: consenters},
		Organizations: ordererOrganizations,
		Policies: map[string]ConfigtxPolicy{
			fconfigtx.ReadersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Readers",
			},
			fconfigtx.WritersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
			fconfigtx.AdminsPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Admins",
			},
			fconfigtx.BlockValidationPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
		},
		Capabilities: map[string]bool{
			"V1_4_2": true,
		},
	}

	var peerOrganizations []ConfigtxOrganization
	peerOrgsPath := filepath.Join(filePath, "crypto-config", "peerOrganizations")
	peerMap := orderPeerOrdererByOrg(peers)
	for _, peerNodes := range peerMap {
		rand.Seed(time.Now().UnixNano())
		peerIndex := rand.Intn(len(peerNodes))
		randomPeer := peerNodes[peerIndex]
		var anchorPeers []ConfigtxAnchorPeer
		anchorPeer := ConfigtxAnchorPeer{
			Host: randomPeer.GetHostname(),
			Port: uint32(randomPeer.NodePort),
		}
		anchorPeers = append(anchorPeers, anchorPeer)
		peerOrganization := ConfigtxOrganization{
			Name:   randomPeer.OrgId,
			ID:     randomPeer.OrgId,
			MSPDir: filepath.Join(peerOrgsPath, randomPeer.Domain, "msp"),
			Policies: map[string]ConfigtxPolicy{
				fconfigtx.ReadersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.peer', '%[1]s.client')", randomPeer.OrgId),
				},
				fconfigtx.WritersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.client')", randomPeer.OrgId),
				},
				fconfigtx.AdminsPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%s.admin')", randomPeer.OrgId),
				},
			},
			AnchorPeers: anchorPeers,
		}
		// TODO 2.0
		peerOrganizations = append(peerOrganizations, peerOrganization)
	}

	application := ConfigtxApplication{
		Organizations: peerOrganizations,
		Policies: map[string]ConfigtxPolicy{
			fconfigtx.ReadersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Readers",
			},
			fconfigtx.WritersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
			fconfigtx.AdminsPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Admins",
			},
			// TODO 2.0
		},
		Capabilities: map[string]bool{
			"V1_4_2": true,
			"V1_3":   false,
			"V1_2":   false,
			"V1_1":   false,
		},
	}

	switch ordererType {
	case ordererType_ETCDRAFT:
		application.Organizations = ordererOrganizations
		configtx.Profiles = map[string]ConfigtxProfile{
			defaultGenesisName: {
				Orderer:     orderer,
				Application: application,
				Capabilities: map[string]bool{
					"V1_4_3": true,
					"V1_3":   false,
					"V1_1":   false,
				},
				Consortiums: map[string]ConfigtxConsortium{
					defaultConsortiumName: {
						Organizations: peerOrganizations,
					},
				},
				Policies: map[string]ConfigtxPolicy{
					fconfigtx.ReadersPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Readers",
					},
					fconfigtx.WritersPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Writers",
					},
					fconfigtx.AdminsPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Admins",
					},
				},
			},
		}
	default:
		configtx.Profiles = map[string]ConfigtxProfile{
			defaultGenesisName: {
				Orderer: orderer,
				Capabilities: map[string]bool{
					"V1_4_3": true,
					"V1_3":   false,
					"V1_1":   false,
				},
				Consortiums: map[string]ConfigtxConsortium{
					defaultConsortiumName: {
						Organizations: peerOrganizations,
					},
				},
				Policies: map[string]ConfigtxPolicy{
					fconfigtx.ReadersPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Readers",
					},
					fconfigtx.WritersPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Writers",
					},
					fconfigtx.AdminsPolicyKey: {
						Type: fconfigtx.ImplicitMetaPolicyType,
						Rule: "ANY Admins",
					},
				},
			},
		}
	}

	configtx.Profiles[defaultChannelProfileName] = ConfigtxProfile{
		Consortium:  defaultConsortiumName,
		Application: application,
		Capabilities: map[string]bool{
			"V1_4_3": true,
			"V1_3":   false,
			"V1_1":   false,
		},
		Policies: map[string]ConfigtxPolicy{
			fconfigtx.ReadersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Readers",
			},
			fconfigtx.WritersPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
			fconfigtx.AdminsPolicyKey: {
				Type: fconfigtx.ImplicitMetaPolicyType,
				Rule: "ANY Admins",
			},
		},
	}
	data, err := yaml.Marshal(configtx)
	if err != nil {
		return err
	}

	path := filepath.Join(filePath, defaultConfigtxFileName)
	return ioutil.WriteFile(path, data, 0755)
}

// GenerateLocallyTestNetworkConfigtx 生成一个用于本地测试使用的configtx.yaml文件
func GenerateLocallyTestNetworkConfigtx(filePath string) error {
	var configtx Configtx
	ordererOrg := ConfigtxOrganization{
		Name:   "OrdererOrg",
		ID:     "OrdererMSP",
		MSPDir: "crypto-config/ordererOrganizations/example.com/msp",
		Policies: map[string]ConfigtxPolicy{
			"Readers": {
				Type: "Signature",
				Rule: "OR('OrdererMSP.member')",
			},
			"Writers": {
				Type: "Signature",
				Rule: "OR('OrdererMSP.member')",
			},
			"Admins": {
				Type: "Signature",
				Rule: "OR('OrdererMSP.admin')",
			},
		},
	}

	org1 := ConfigtxOrganization{
		Name:   "Org1MSP",
		ID:     "Org1MSP",
		MSPDir: "crypto-config/peerOrganizations/org1.example.com/msp",
		Policies: map[string]ConfigtxPolicy{
			"Readers": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin', 'Org1MSP.peer', 'Org1MSP.client')",
			},
			"Writers": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin', 'Org1MSP.client')",
			},
			"Admins": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin')",
			},
		},
		AnchorPeers: []ConfigtxAnchorPeer{
			{
				Host: "peer0.org1.example.com",
				Port: 7051,
			},
		},
	}

	org2 := ConfigtxOrganization{
		Name:   "Org2MSP",
		ID:     "Org2MSP",
		MSPDir: "crypto-config/peerOrganizations/org2.example.com/msp",
		Policies: map[string]ConfigtxPolicy{
			"Readers": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin', 'Org2MSP.peer', 'Org2MSP.client')",
			},
			"Writers": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin', 'Org2MSP.client')",
			},
			"Admins": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin')",
			},
		},
		AnchorPeers: []ConfigtxAnchorPeer{
			{
				Host: "peer0.org2.example.com",
				Port: 9051,
			},
		},
	}

	twoOrgsOrdererGenesis := ConfigtxProfile{
		Orderer: ConfigtxOrderer{
			OrdererType:  "solo",
			Addresses:    []string{"orderer.example.com:7050"},
			BatchTimeout: "2s",
			BatchSize: ConfigtxBatchSize{
				MaxMessageCount:   10,
				AbsoluteMaxBytes:  "99 MB",
				PreferredMaxBytes: "512 KB",
			},
			Kafka: ConfigtxKafka{
				Brokers: []string{"127.0.0.1:9092"},
			},
			EtcdRaft: ConfigtxEtcdRaft{
				Consenters: []ConfigtxConsenter{
					{
						Host:          "orderer.example.com",
						Port:          7050,
						ClientTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt",
						ServerTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt",
					},
					{
						Host:          "orderer2.example.com",
						Port:          7050,
						ClientTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer2.example.com/tls/server.crt",
						ServerTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer2.example.com/tls/server.crt",
					},
					{
						Host:          "orderer3.example.com",
						Port:          7050,
						ClientTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer3.example.com/tls/server.crt",
						ServerTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer3.example.com/tls/server.crt",
					},
					{
						Host:          "orderer4.example.com",
						Port:          7050,
						ClientTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer4.example.com/tls/server.crt",
						ServerTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer4.example.com/tls/server.crt",
					},
					{
						Host:          "orderer5.example.com",
						Port:          7050,
						ClientTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer5.example.com/tls/server.crt",
						ServerTLSCert: "crypto-config/ordererOrganizations/example.com/orderers/orderer5.example.com/tls/server.crt",
					},
				},
			},
			Policies: map[string]ConfigtxPolicy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Admins",
				},
				"BlockValidation": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
			Organizations: []ConfigtxOrganization{ordererOrg},
			Capabilities: map[string]bool{
				"V1_4_2": true,
				"V1_1":   false,
			},
		},
		Capabilities: map[string]bool{
			"V1_4_3": true,
			"V1_3":   false,
			"V1_1":   false,
		},
		Consortiums: map[string]ConfigtxConsortium{
			"SampleConsortium": {Organizations: []ConfigtxOrganization{org1, org2}},
		},
		Policies: map[string]ConfigtxPolicy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
		},
	}
	configtx.Profiles = make(map[string]ConfigtxProfile)
	configtx.Profiles["TwoOrgsOrdererGenesis"] = twoOrgsOrdererGenesis

	// 生成channel文件相关配置
	twoOrgsChannel := ConfigtxProfile{
		Consortium: "SampleConsortium",
		Application: ConfigtxApplication{
			Organizations: []ConfigtxOrganization{org1, org2},
			Policies: map[string]ConfigtxPolicy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Admins",
				},
			},
			Capabilities: map[string]bool{
				"V1_4_2": true,
				"V1_3":   false,
				"V1_2":   false,
				"V1_1":   false,
			},
		},
		Capabilities: map[string]bool{
			"V1_4_3": true,
			"V1_3":   false,
			"V1_1":   false,
		},
		Policies: map[string]ConfigtxPolicy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
		},
	}
	configtx.Profiles["TwoOrgsChannel"] = twoOrgsChannel
	path := filepath.Join(filePath, defaultConfigtxFileName)
	data, err := yaml.Marshal(configtx)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0755)
}

func GenerateGensisBlockAndChannelTxAndAnchorPeer(fileDir string, channelId string, nc *network.NetworkConfig) error {
	// generate genesis.block
	configtxgenPath := filepath.Join("tools", "configtxgen")
	channelArtifactsPath := filepath.Join(fileDir, "channel-artifacts")
	if err := os.MkdirAll(channelArtifactsPath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory, path=%s", channelArtifactsPath)
	}

	logger.Infof("begin to generate fabric genesis.block, genesis channel name is %s", defaultGenesisChannel)
	var args []string
	// ./configtxgen -profile TwoOrgsOrdererGenesis -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block
	args = []string{
		fmt.Sprintf("--configPath=%s", fileDir),
		fmt.Sprintf("--profile=%s", defaultGenesisName),
		fmt.Sprintf("--channelID=%s", defaultGenesisChannel),
		fmt.Sprintf("--outputBlock=%s", filepath.Join(channelArtifactsPath, "genesis.block")),
	}
	if err := utils.RunLocalCmd(configtxgenPath, args...); err != nil {
		return err
	}

	// generate channel transaction
	// ./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
	logger.Infof("begin to generate channel.tx, channel=%s", channelId)
	args = []string{
		fmt.Sprintf("--configPath=%s", fileDir),
		fmt.Sprintf("--profile=%s", defaultChannelProfileName),
		fmt.Sprintf("--channelID=%s", channelId),
		fmt.Sprintf("--outputCreateChannelTx=%s", filepath.Join(channelArtifactsPath, fmt.Sprintf("%s.tx", channelId))),
	}
	if err := utils.RunLocalCmd(configtxgenPath, args...); err != nil {
		return err
	}

	peerNodes, _, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return nil
	}
	peerOrgs := make(map[string]interface{})
	for _, node := range peerNodes {
		peerOrgs[node.OrgId] = nil
	}
	// TODO: does fabric2.x need this step???
	// generate anchor peer transaction
	// ./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1anchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
	for org, _ := range peerOrgs {
		logger.Infof("begin to generate anchors.tx, org=%s", org)
		args = []string{
			fmt.Sprintf("--configPath=%s", fileDir),
			fmt.Sprintf("--profile=%s", defaultChannelProfileName),
			fmt.Sprintf("--channelID=%s", channelId),
			fmt.Sprintf("--outputAnchorPeersUpdate=%s", filepath.Join(channelArtifactsPath, fmt.Sprintf("%sanchors.tx", org))),
			fmt.Sprintf("--asOrg=%s", org),
		}
		if err := utils.RunLocalCmd(configtxgenPath, args...); err != nil {
			return err
		}
	}
	return nil
}
