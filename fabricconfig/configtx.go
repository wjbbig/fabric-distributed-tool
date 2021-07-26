package fabricconfig

import (
	"fmt"
	fconfigtx "github.com/hyperledger/fabric-config/configtx"
	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
func orderPeerOrdererByOrg(urls []string) map[string][]string {
	orderedUrl := make(map[string][]string)
	for _, url := range urls {
		_, org, _ := utils.SplitNameOrgDomain(url)
		orderedUrl[org] = append(orderedUrl[org], url)
	}
	return orderedUrl
}

func GenerateConfigtxFile(filePath string, ordererType string, orderers, peers []string) error {
	logger.Infof("begin to generate configtx.yaml, orderer type=%s", ordererType)
	defer logger.Info("finish generating configtx.yaml")
	var configtx Configtx
	var consenters []ConfigtxConsenter
	var ordererOrganizations []ConfigtxOrganization
	ordererOrgsPath := filepath.Join(filePath, "crypto-config", "ordererOrganizations")
	ordererMap := orderPeerOrdererByOrg(orderers)
	for _, ordererUrls := range ordererMap {
		for _, url := range ordererUrls {
			ordererArgs := strings.Split(url, ":")
			_, _, ordererDomain := utils.SplitNameOrgDomain(ordererArgs[0])
			serverCertPath := filepath.Join(ordererOrgsPath, ordererDomain, "orderers", ordererArgs[0], "tls/server.crt")
			port, err := strconv.Atoi(ordererArgs[1])
			if err != nil {
				return errors.Wrap(err, "get orderer port failed")
			}
			consenter := ConfigtxConsenter{
				Host:          ordererArgs[0],
				Port:          uint32(port),
				ClientTLSCert: serverCertPath,
				ServerTLSCert: serverCertPath,
			}
			consenters = append(consenters, consenter)
		}
		ordererArgs := strings.Split(ordererUrls[0], ":")
		_, ordererOrgName, ordererDomain := utils.SplitNameOrgDomain(ordererArgs[0])

		ordererOrganization := ConfigtxOrganization{
			Name:   ordererOrgName,
			ID:     ordererOrgName,
			MSPDir: filepath.Join(ordererOrgsPath, ordererDomain, "msp"),
			Policies: map[string]ConfigtxPolicy{
				fconfigtx.ReadersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin','%[1]s.orderer','%[1]s.client')", ordererOrgName),
				},
				fconfigtx.WritersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.orderer', '%[1]s.client')", ordererOrgName),
				},
				fconfigtx.AdminsPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%s.admin')", ordererOrgName),
				},
			},
		}
		ordererOrganizations = append(ordererOrganizations, ordererOrganization)
	}

	orderer := ConfigtxOrderer{
		OrdererType:  ordererType,
		Addresses:    orderers,
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
	for _, peerUrls := range peerMap {
		rand.Seed(time.Now().UnixNano())
		peerIndex := rand.Intn(len(peerUrls))
		peerArgs := strings.Split(peerUrls[peerIndex], ":")
		port, err := strconv.Atoi(peerArgs[1])
		if err != nil {
			return err
		}
		var anchorPeers []ConfigtxAnchorPeer
		anchorPeer := ConfigtxAnchorPeer{
			Host: peerArgs[0],
			Port: uint32(port),
		}
		anchorPeers = append(anchorPeers, anchorPeer)
		_, org, domain := utils.SplitNameOrgDomain(peerArgs[0])
		peerOrganization := ConfigtxOrganization{
			Name:   org,
			ID:     org,
			MSPDir: filepath.Join(peerOrgsPath, domain, "msp"),
			Policies: map[string]ConfigtxPolicy{
				fconfigtx.ReadersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.peer', '%[1]s.client')", org),
				},
				fconfigtx.WritersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.client')", org),
				},
				fconfigtx.AdminsPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%s.admin')", org),
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
				Rule: fmt.Sprintf("ANY Readers"),
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

func GenerateGensisBlockAndChannelTxAndAnchorPeer(fileDir string, channelId string, peerOrgs []string) error {
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

	// TODO: does fabric2.x need this step???
	// generate anchor peer transaction
	// ./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1anchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
	for _, org := range peerOrgs {
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
