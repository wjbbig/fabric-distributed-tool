package network

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHomeDirConfig(t *testing.T) {
	_, err := Load()
	require.Error(t, err)

	config := NewHomeDirConfig()
	config.AddNetwork("testnetwork", "/data/fdt")
	err = config.Store()
	require.NoError(t, err)

	config, err = Load()
	require.NoError(t, err)

	t.Log(config.GetNetworkPath("testnetwork"))

}
