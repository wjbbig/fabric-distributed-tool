package sdkutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const connectionFilePath = "/opt/fdt/connection-config.yaml"

func TestFabricSDKDriver_CreateChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.CreateChannel("mychannel", "testpeerorg1",
		"/opt/fdt", "orderer.testordererorg", "peer.testpeerorg1")
	require.NoError(t, err)
}

func TestFabricSDKDriver_JoinChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.JoinChannel("mychannel", "testpeerorg2", "orderer.testordererorg", "peer.testpeerorg2")
	require.NoError(t, err)
	err = sdk.JoinChannel("mychannel", "testpeerorg1", "orderer.testordererorg", "peer.testpeerorg1")
	require.NoError(t, err)
}

func TestFabricSDKDriver_InstallCC(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()
	err = sdk.InstallCC("mycc", "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go",
		"v0.1", "mychannel", "testpeerorg1", "peer.testpeerorg1")
	require.NoError(t, err)
	err = sdk.InstallCC("mycc", "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go",
		"v0.1", "mychannel", "testpeerorg2", "peer.testpeerorg2")
	require.NoError(t, err)
}

func TestFabricSDKDriver_InstantiateCC(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.InstantiateCC("mycc", "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go",
		"v0.1", "mychannel", "testpeerorg2", "OR('testpeerorg1.peer','testpeerorg2.peer')",
		"peer.testpeerorg2", []string{"init", "a", "100", "b", "20"})
	require.NoError(t, err)
}
