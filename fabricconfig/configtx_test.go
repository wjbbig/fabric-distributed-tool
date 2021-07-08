package fabricconfig

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateLocallyTestNetworkConfigtx(t *testing.T) {
	err := GenerateLocallyTestNetworkConfigtx("./")
	require.NoError(t, err)
}

func TestGenerateConfigtxFile(t *testing.T) {
	err := GenerateConfigtxFile("./", "solo", []string{"orderer.example.com:7050"},
		[]string{"peer0.org1.example.com:7051", "peer1.org1.example.com:8051", "peer0.org2.example.com:9051"})
	require.NoError(t, err)
}

func TestGenerateConfigFiles(t *testing.T) {
	err := GenerateCryptoConfigFile("/home/ubuntu/workspace/go/src/github.com/hyperledger/fabric-samples/my-first-network",
		[]string{"peer0.org1.example.com", "peer1.org1.example.com", "peer0.org2.example.com", "peer1.org2.example.com"},
		[]string{"orderer.example.com"})
	require.NoError(t, err)

	err = GenerateConfigtxFile("/home/ubuntu/workspace/go/src/github.com/hyperledger/fabric-samples/my-first-network",
		"solo", []string{"orderer.example.com:7050"},
		[]string{"peer0.org1.example.com:7051", "peer1.org1.example.com:8051",
			"peer0.org2.example.com:9051", "peer1.org2.example.com:10051"})
	require.NoError(t, err)
}
