package fabricconfig

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateLocallyTestNetworkConfigtx(t *testing.T) {
	err := GenerateLocallyTestNetworkConfigtx("./")
	require.NoError(t, err)
}
