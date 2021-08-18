package fabricconfig

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/wjbbig/fabric-distributed-tool/tools/configtxgen/encoder"
	"github.com/wjbbig/fabric-distributed-tool/tools/configtxgen/genesisconfig"
	"github.com/wjbbig/fabric-distributed-tool/tools/configtxlator/update"
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
	Options    EtcdRaftOptions     `yaml:"Options,omitempty"`
}

type EtcdRaftOptions struct {
	TickInterval         string `yaml:"TickInterval,omitempty"`
	ElectionTick         uint32 `yaml:"ElectionTick,omitempty"`
	HeartbeatTick        uint32 `yaml:"HeartbeatTick,omitempty"`
	MaxInflightBlocks    uint32 `yaml:"MaxInflightBlocks,omitempty"`
	SnapshotIntervalSize uint32 `yaml:"SnapshotIntervalSize,omitempty"`
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
	// TODO raft错误
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
		ordererApplication := application
		ordererApplication.Organizations = ordererOrganizations
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

func GenerateGensisBlockAndChannelTxAndAnchorPeerUsingBinary(fileDir string, channelId string, nc *network.NetworkConfig) error {
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
	// generate anchor peer transaction
	// ./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1anchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
	for org := range peerOrgs {
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

func GenerateGenesisBlockAndChannelTxAndAnchorPeer(fileDir string, channelId string, nc *network.NetworkConfig) error {
	// generate genesis.block
	channelArtifactsPath := filepath.Join(fileDir, "channel-artifacts")
	if err := os.MkdirAll(channelArtifactsPath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory, path=%s", channelArtifactsPath)
	}

	logger.Infof("begin to generate fabric genesis.block, genesis channel name is %s", defaultGenesisChannel)
	profileConfig := genesisconfig.Load(defaultGenesisName, fileDir)
	outputBlockPath := filepath.Join(channelArtifactsPath, "genesis.block")
	if err := doOutputBlock(profileConfig, defaultGenesisChannel, outputBlockPath); err != nil {
		return err
	}

	// generate channel transaction
	logger.Infof("begin to generate channel.tx, channel=%s", channelId)
	profileConfig = genesisconfig.Load(defaultChannelProfileName, fileDir)
	outputChannelCreateTxPath := filepath.Join(channelArtifactsPath, fmt.Sprintf("%s.tx", channelId))
	if err := doOutputChannelCreateTx(profileConfig, channelId, outputChannelCreateTxPath); err != nil {
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
	for org := range peerOrgs {
		logger.Infof("begin to generate anchors.tx, org=%s", org)
		profileConfig = genesisconfig.Load(defaultChannelProfileName, fileDir)
		outputAnchorPeersUpdatePath := filepath.Join(channelArtifactsPath, fmt.Sprintf("%sanchors.tx", org))
		if err := doOutputAnchorPeersUpdate(profileConfig, channelId, outputAnchorPeersUpdatePath, org); err != nil {
			return err
		}
	}
	return nil
}

func doOutputBlock(config *genesisconfig.Profile, channelId string, outputBlock string) error {
	pgen, err := encoder.NewBootstrapper(config)
	if err != nil {
		return errors.Wrap(err, "could not create bootstrapper")
	}
	if config.Orderer == nil {
		return errors.Errorf("refusing to generate block which is missing orderer section")
	}
	if config.Consortiums == nil {
		logger.Warn("genesis block does not contain a consortiums group definition.  This block cannot be used for orderer bootstrap")
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelId)
	if err := utils.WriteFile(outputBlock, protoutil.MarshalOrPanic(genesisBlock), 0640); err != nil {
		return fmt.Errorf("error writing genesis block: %s", err)
	}
	return nil
}

func doOutputChannelCreateTx(conf *genesisconfig.Profile, channelId string, outputChannelCreateTx string) error {
	configtx, err := encoder.MakeChannelCreationTransaction(channelId, nil, conf)
	if err != nil {
		return err
	}
	if err := utils.WriteFile(outputChannelCreateTx, protoutil.MarshalOrPanic(configtx), 0640); err != nil {
		return fmt.Errorf("error writing channel create tx: %s", err)
	}
	return nil
}

func doOutputAnchorPeersUpdate(conf *genesisconfig.Profile, channelId, outputAnchorPeersUpdate, asOrg string) error {
	if asOrg == "" {
		return fmt.Errorf("must specify an organization to update the anchor peer for")
	}

	if conf.Application == nil {
		return fmt.Errorf("cannot update anchor peers without an application section")
	}

	original, err := encoder.NewChannelGroup(conf)
	if err != nil {
		return errors.WithMessage(err, "error parsing profile as channel group")
	}
	original.Groups[channelconfig.ApplicationGroupKey].Version = 1

	updated := proto.Clone(original).(*cb.ConfigGroup)

	originalOrg, ok := original.Groups[channelconfig.ApplicationGroupKey].Groups[asOrg]
	if !ok {
		return errors.Errorf("org with name '%s' does not exist in config", asOrg)
	}

	if _, ok = originalOrg.Values[channelconfig.AnchorPeersKey]; !ok {
		return errors.Errorf("org '%s' does not have any anchor peers defined", asOrg)
	}

	delete(originalOrg.Values, channelconfig.AnchorPeersKey)

	updt, err := update.Compute(&cb.Config{ChannelGroup: original}, &cb.Config{ChannelGroup: updated})
	if err != nil {
		return errors.WithMessage(err, "could not compute update")
	}
	updt.ChannelId = channelId

	newConfigUpdateEnv := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: protoutil.MarshalOrPanic(updt),
	}

	updateTx, err := protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelId, nil, newConfigUpdateEnv, 0, 0)
	if err != nil {
		return errors.WithMessage(err, "could not create signed envelope")
	}
	err = utils.WriteFile(outputAnchorPeersUpdate, protoutil.MarshalOrPanic(updateTx), 0640)
	if err != nil {
		return fmt.Errorf("Error writing channel anchor peer update: %s", err)
	}
	return nil
}
