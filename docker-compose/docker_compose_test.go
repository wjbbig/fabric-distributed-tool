package docker_compose

import (
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

func TestDetectImageNameAndTag(t *testing.T) {
	tag, err := detectImageNameAndTag("orderer")
	require.NoError(t, err)
	t.Log(tag)
}

func TestGeneratePeerDockerComposeFile(t *testing.T) {
	err := GeneratePeerDockerComposeFile(".", "peer0.org1.example.com:7050", "peer0.org1.example.com:7050", nil)
	require.NoError(t, err)
}

func TestBasePath(t *testing.T) {
	base := path.Base("/opt/fdt")
	t.Log(base)
}
