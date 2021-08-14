package sshutil

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"os"
	"testing"
)

func TestSSH(t *testing.T) {
	client, err := ssh.Dial("tcp", "101.91.231.36:22", &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("Blockchain@123")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	require.NoError(t, err)
	defer client.Close()

	session, err := client.NewSession()
	require.NoError(t, err, "create ssh session failed")
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run("docker ps")
	require.NoError(t, err, "run cmd failed")
}

func TestSshClient_RunCmd(t *testing.T) {
	localCli, err := newSSHClient("ubuntu", "123456", "127.0.0.1:22", "peer", false)
	require.NoError(t, err)
	defer localCli.Close()
	require.Equal(t, localCli.local, true)
	remoteCli, err := newSSHClient("root", "Blockchain@123", "101.91.231.36:22", "peer", false)
	require.NoError(t, err)
	defer remoteCli.Close()
	require.Equal(t, remoteCli.local, false)
	t.Log("------local------")
	err = localCli.RunCmd("docker ps")
	require.NoError(t, err)
	t.Log("------remote-----")
	err = remoteCli.RunCmd("docker ps")
	require.NoError(t, err)
}

func TestSshClient_Sftp(t *testing.T) {
	remoteCli, err := newSSHClient("root", "Blockchain@123", "101.91.231.36:22", "peer", false)
	require.NoError(t, err)
	defer remoteCli.Close()
	//err = remoteCli.Sftp("/opt/db/cache.db", "/root")
	//require.NoError(t, err)

	// 测试传输文件夹
	err = remoteCli.Sftp("/home/ubuntu/workspace/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config",
		"/root/crypto-config")
	require.NoError(t, err)
}
