package network

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateNetworkConfig(t *testing.T) {
	_, err := GenerateNetworkConfig(".", "testnetwork","mychannel","etcd/raft","mycc","github.com/hyperledger/fabric-sample/go/","v1","init,1,2,3",
		"Or('org1.peer','org2.peer')",[]string{"peer0.org1.example.com:7051@user@127.0.0.1:22:32432"}, []string{"orderer0.orderer.example.com:7050@user@127.0.0.1:22:32432"})
	require.NoError(t, err)
}