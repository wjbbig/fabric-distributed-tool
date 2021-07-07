package fabricconfig

import (
	"fmt"
	fconfigtx "github.com/hyperledger/fabric-config/configtx"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultConfigtxFileName = "configtx.yaml"

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

func GenerateConfigtxFile(filePath string, orderers, peers []string) error {
	var configtx Configtx
	var consenters []ConfigtxConsenter
	var ordererOrganizations []ConfigtxOrganization
	ordererOrgsPath := filepath.Join(filePath, "crypto-config", "ordererOrganizations")
	for _, orderer := range orderers {
		ordererArgs := strings.Split(orderer, ":")
		ordererName, ordererOrgName, ordererDomain := splitNameOrgDomain(ordererArgs[0])
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
		mspId := fmt.Sprintf("%sMSP", ordererOrgName)
		ordererOrganization := ConfigtxOrganization{
			Name:   ordererOrgName,
			ID:     mspId,
			MSPDir: filepath.Join(ordererOrgsPath, ordererDomain, "orderers", ordererArgs[0], "msp/"),
			Policies: map[string]ConfigtxPolicy{
				fconfigtx.ReadersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin','%[1]s.orderer','%[1]s.client')", mspId),
				},
				fconfigtx.WritersPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%[1]s.admin', '%[1]s.orderer', '%[1]s.client')", mspId),
				},
				fconfigtx.AdminsPolicyKey: {
					Type: fconfigtx.SignaturePolicyType,
					Rule: fmt.Sprintf("OR('%s.admin')", mspId),
				},
			},
		}
		ordererOrganizations = append(ordererOrganizations, ordererOrganization)
	}

	return nil
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
