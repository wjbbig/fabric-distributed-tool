package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDoNewOrgPeerJoinChannel(t *testing.T) {
	err := DoNewOrgPeerJoinChannel("/opt/fdt", "mychannel", "peer.testpeerorg3")
	require.NoError(t, err)
}
