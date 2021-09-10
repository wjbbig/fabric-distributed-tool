package sdkutil

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const connectionFilePath = "/opt/fdt/connection-config.yaml"

func TestFabricSDKDriver_CreateChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.CreateChannel("mychannel", "testpeerorg1",
		"/opt/fdt", "orderer1.testordererorg", "peer.testpeerorg1")
	require.NoError(t, err)
}

func TestFabricSDKDriver_JoinChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.JoinChannel("mychannel", "testpeerorg2", "orderer1.testordererorg", "peer.testpeerorg2")
	require.NoError(t, err)
	err = sdk.JoinChannel("mychannel", "testpeerorg1", "orderer1.testordererorg", "peer.testpeerorg1")
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

func TestFabricSDKDriverV2(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()
	err = sdk.CreateChannel("mychannel", "testpeerorg1",
		"/opt/fdt", "orderer1.testordererorg", "peer.testpeerorg1")
	require.NoError(t, err)
	err = sdk.JoinChannel("mychannel", "testpeerorg2", "orderer1.testordererorg", "peer.testpeerorg2")
	require.NoError(t, err)
	err = sdk.JoinChannel("mychannel", "testpeerorg1", "orderer1.testordererorg", "peer.testpeerorg1")
	require.NoError(t, err)
	_, ccPkg, err := sdk.PackageCC("mycc", "/home/wjbbig/workspace/go/src/github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go",
		"/opt/fdt/chaincode_packages/mycc.tar.gz")
	require.NoError(t, err)

	packageId, err := sdk.InstallCCV2("mycc", "mychannel", "testpeerorg1", "peer.testpeerorg1", ccPkg)
	require.NoError(t, err)

	_, err = sdk.InstallCCV2("mycc", "mychannel", "testpeerorg2", "peer.testpeerorg2", ccPkg)
	require.NoError(t, err)

	ccPkg1, err := sdk.QueryGetInstalled("mychannel", "testpeerorg1", packageId, "peer.testpeerorg1")
	require.NoError(t, err)
	require.Equal(t, ccPkg1, ccPkg)

	err = sdk.ApproveCC("mycc", "v1", "AND('testpeerorg1.peer','testpeerorg2.peer')", "mychannel",
		"testpeerorg1", "peer.testpeerorg1", "orderer5.testordererorg", packageId, 1, true)
	require.NoError(t, err)
	time.Sleep(time.Second)
	err = sdk.QueryApprovedCC("mycc", "mychannel", "testpeerorg1", "peer.testpeerorg1", 1)
	require.NoError(t, err)
	err = sdk.ApproveCC("mycc", "v1", "AND('testpeerorg1.peer','testpeerorg2.peer')", "mychannel",
		"testpeerorg2", "peer.testpeerorg2", "orderer5.testordererorg", packageId, 1, true)
	require.NoError(t, err)
	time.Sleep(time.Second)
	err = sdk.QueryApprovedCC("mycc", "mychannel", "testpeerorg2", "peer.testpeerorg2", 1)
	require.NoError(t, err)
	time.Sleep(time.Second)
	err = sdk.CommitCC("mycc", "v1", "AND('testpeerorg1.peer','testpeerorg2.peer')", "mychannel",
		"testpeerorg1", "peer.testpeerorg1", "orderer5.testordererorg", 1, true)
	require.NoError(t, err)
	err = sdk.InitCC("mycc", "mychannel", "testpeerorg2", "InitLedger", []string{}, []string{"peer.testpeerorg1", "peer.testpeerorg2"})
	require.NoError(t, err)
}

func TestGetConfigBlock(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	block, err := sdk.getConfigBlock("mychannel", "testordererorg", "orderer.testordererorg")
	require.NoError(t, err)

	t.Log(block.Header.Number)
	t.Log(len(block.Data.Data))

	config, err := getChannelConfigFromBlock(block)
	require.NoError(t, err)
	require.NotNil(t, config)

	groups := config.ChannelGroup.Groups["Application"].Groups
	for _, group := range groups {
		fmt.Println(group)
	}
}
