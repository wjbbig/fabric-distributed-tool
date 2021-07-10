package sshutil

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateSSHConfig(t *testing.T) {
	err := GenerateSSHConfig(".", []string{
		"peer0.org1.example.com:7050@root@127.0.0.1:22:password",
		"peer8.org2.example.com:7051@ubuntu@127.0.0.1:1022:passw:ord",
	})

	require.NoError(t, err)
}

func TestUnmarshalSSHConfig(t *testing.T) {
	sshConfig, err := UnmarshalSSHConfig(".")
	require.NoError(t, err)

	t.Log(sshConfig)
}
