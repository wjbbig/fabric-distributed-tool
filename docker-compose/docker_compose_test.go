package docker_compose

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDetectImageNameAndTag(t *testing.T) {
	tag, err := detectImageNameAndTag("orderer")
	require.NoError(t, err)
	t.Log(tag)
}
