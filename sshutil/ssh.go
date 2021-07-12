package sshutil

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type SSHUtil struct {
	remoteClients map[string]*sshClient
}

func NewSSHUtil() *SSHUtil {
	return &SSHUtil{make(map[string]*sshClient)}
}

func (su *SSHUtil) Add(peerName, username, password, address string) error {
	cli, err := NewSSHClient(username, password, address)
	if err != nil {
		return err
	}
	su.remoteClients[peerName] = cli
	return nil
}

type sshClient struct {
	username string
	password string
	address  string
	local    bool
	client   *ssh.Client
}

func NewSSHClient(username, password, address string) (*sshClient, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	var local bool
	if addr.IP.IsLoopback() || addr.IP.IsUnspecified() {
		local = true
	}

	cli := &sshClient{
		username: username,
		password: password,
		address:  addr.String(),
		local:    local,
	}

	if !local {
		if cli.client, err = ssh.Dial("tcp", cli.address, &ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

// RunCmd 执行命令
func (cli *sshClient) RunCmd(cmd string) error {
	var buffer bytes.Buffer
	if cli.local {
		args := strings.Split(cmd, " ")
		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = &buffer
		if err := command.Run(); err != nil {
			return err
		}
		if buffer.Len() != 0 {
			return errors.Errorf("run command failed, address=%s, err=%s", cli.address, buffer.String())
		}
	} else {
		session, err := cli.client.NewSession()
		if err != nil {
			return errors.Wrapf(err, "create ssh session failed, address=%s", cli.address)
		}
		defer session.Close()
		session.Stdout = os.Stdout
		session.Stderr = &buffer
		if err = session.Run(cmd); err != nil {
			return err
		}
		if buffer.Len() != 0 {
			return errors.Errorf("run command failed, address=%s, err=%s", cli.address, buffer.String())
		}
	}
	return nil
}

// Sftp 传输文件
func (cli *sshClient) Sftp(localFilePath string, remoteDir string) error {
	// 本地节点不用移动文件
	if cli.local {
		return nil
	}
	sftpClient, err := sftp.NewClient(cli.client)
	if err != nil {
		return errors.Wrap(err, "start sftp client failed")
	}
	defer sftpClient.Close()

	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return transferDir(localFilePath, remoteDir, sftpClient)
	}
	return transferFile(localFilePath, remoteDir, sftpClient)
}

func transferDir(localDirPath string, remoteDir string, sftpClient *sftp.Client) error {
	dir, err := ioutil.ReadDir(localDirPath)
	if err != nil {
		return errors.Wrapf(err, "read local dir failed, path=%s", localDirPath)
	}
	for _, file := range dir {
		localFile := filepath.Join(localDirPath, file.Name())
		remoteFile := filepath.Join(remoteDir, file.Name())
		if file.IsDir() {
			if err := sftpClient.MkdirAll(localFile); err != nil {
				return errors.Wrapf(err, "failed to mkdir, path=%s", localFile)
			}
			if err := transferDir(localFile, remoteFile, sftpClient); err != nil {
				return errors.Wrapf(err, "failed to deep transferDir, path=%s", localFile)
			}
		} else {
			if err := transferFile(localFile, remoteDir, sftpClient); err != nil {
				return err
			}
		}
	}

	return nil
}

// transferFile 用于传输单个文件
func transferFile(localFilePath string, remoteDir string, sftpClient *sftp.Client) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return errors.Wrapf(err, "open file failed, filepath=%s", localFilePath)
	}
	defer localFile.Close()

	var remoteFileName = path.Base(localFilePath)
	remoteFileNamePath := filepath.Join(remoteDir, remoteFileName)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return errors.Wrapf(err, "create dir %s failed", remoteDir)
	}
	dstFile, err := sftpClient.Create(remoteFileNamePath)
	if err != nil {
		return errors.Wrapf(err, "create remote file failed, filepath=%s", remoteFileName)
	}
	defer dstFile.Close()

	buf := make([]byte, 40960)
	for {
		n, _ := localFile.Read(buf)
		if n == 0 {
			break
		}
		_, err := dstFile.Write(buf)
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to write remote file")
		}
	}
	return nil
}

func (cli *sshClient) Close() error {
	if cli.client != nil {
		return cli.client.Close()
	}
	return nil
}
