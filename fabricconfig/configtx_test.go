package fabricconfig

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestGenerateLocallyTestNetworkConfigtx(t *testing.T) {
	err := GenerateLocallyTestNetworkConfigtx("./")
	require.NoError(t, err)
}

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

// func TestGenerateConfigtxFile(t *testing.T) {
//	err := GenerateConfigtxFile("./", "solo", []string{"orderer.example.com:7050"},
//		[]string{"peer0.org1.example.com:7051", "peer1.org1.example.com:8051", "peer0.org2.example.com:9051"})
//	require.NoError(t, err)
//}
//
//func TestGenerateConfigFiles(t *testing.T) {
//	err := GenerateCryptoConfigFile("/home/ubuntu/workspace/go/src/github.com/hyperledger/fabric-samples/my-first-network",
//		[]string{"peer0.org1.example.com", "peer1.org1.example.com", "peer0.org2.example.com", "peer1.org2.example.com"},
//		[]string{"orderer.example.com"})
//	require.NoError(t, err)
//
//	err = GenerateConfigtxFile("/home/ubuntu/workspace/go/src/github.com/hyperledger/fabric-samples/my-first-network",
//		"solo", []string{"orderer.example.com:7050"},
//		[]string{"peer0.org1.example.com:7051", "peer1.org1.example.com:8051",
//			"peer0.org2.example.com:9051", "peer1.org2.example.com:10051"})
//	require.NoError(t, err)
//}

func TestGenerateConfigGroup(t *testing.T) {
	dataDir := "/opt/fdt"
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	require.NoError(t, err)
	newOrgPeer := nc.GetNode("peer.testpeerorg3")

	org := GenerateOrgProfile(dataDir, newOrgPeer, nc.Version)

	configGroup, err := GenerateConfigGroup(org)
	require.NoError(t, err)
	fmt.Println(configGroup)
}
