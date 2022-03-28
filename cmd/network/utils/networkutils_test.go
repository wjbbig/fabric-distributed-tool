package utils

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestDoNewOrgPeerJoinChannel(t *testing.T) {
	err := DoNewOrgPeerJoinChannel("/opt/fdt", "mychannel", "peer.testpeerorg3")
	require.NoError(t, err)
}

func TestDoCreateCANodeForOrg(t *testing.T) {
	dataDir, err := filepath.Abs("../../../fdtdata")
	require.NoError(t, err)

	err = DoCreateCANodeForOrg(dataDir, "testpeerorg1", "admin", "adminpw")
	require.NoError(t, err)
}
