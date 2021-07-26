package sdkutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const connectionFilePath = "/opt/fdt/connection-config.yaml"

func TestCreateChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.CreateChannel("mychannel", "testpeerorg", "/opt/fdt", "orderer.testordererorg")
	require.NoError(t, err)
}

func TestJoinChannel(t *testing.T) {
	sdk, err := NewFabricSDKDriver(connectionFilePath)
	require.NoError(t, err)
	defer sdk.Close()

	err = sdk.JoinChannel("mychannel", "testpeerorg", "orderer.testordererorg")
	require.NoError(t, err)
}
