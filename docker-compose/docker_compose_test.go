package docker_compose

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wjbbig/fabric-distributed-tool/utils"
)

func TestDetectImageNameAndTag(t *testing.T) {
	tag, err := utils.DetectImageNameAndTag("orderer")
	require.NoError(t, err)
	t.Log(tag)
}

func TestGeneratePeerDockerComposeFile(t *testing.T) {
	err := GeneratePeerDockerComposeFile(".", "peer0.org1.example.com:7050", "peer0.org1.example.com:7050", nil, false)
	require.NoError(t, err)
}

func TestBasePath(t *testing.T) {
	base := path.Base("/opt/fdt")
	t.Log(base)
}
