package sshutil

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type SSHUtil struct {
	remoteClients map[string]*SSHClient
}

func NewSSHUtil() *SSHUtil {
	return &SSHUtil{make(map[string]*SSHClient)}
}

func (su *SSHUtil) Add(peerName, username, password, address, nodeType, locPath, desPath string, needCouchdb bool) error {
	cli, err := newSSHClient(username, password, address, nodeType, locPath, desPath, needCouchdb)
	if err != nil {
		return err
	}
	su.remoteClients[peerName] = cli
	return nil
}

func (su *SSHUtil) Clients() map[string]*SSHClient {
	return su.remoteClients
}

func (su *SSHUtil) GetClientByName(name string) *SSHClient {
	return su.remoteClients[name]
}

func (su *SSHUtil) CloseAll() {
	for _, client := range su.remoteClients {
		client.Close()
	}
}

type SSHClient struct {
	Username  string
	Password  string
	Address   string
	LocalPath string
	DestPath  string
	NodeType  string
	NeedCouch bool
	local     bool
	client    *ssh.Client
}

func newSSHClient(username, password, address, nodeType, locPath, desPath string, needCouchdb bool) (*SSHClient, error) {
	local, err := utils.CheckLocalIp(address)
	if err != nil {
		return nil, err
	}
	cli := &SSHClient{
		Username:  username,
		Password:  password,
		Address:   address,
		NodeType:  nodeType,
		NeedCouch: needCouchdb,
		local:     local,
		LocalPath: locPath,
		DestPath:  desPath,
	}

	if !local {
		if cli.client, err = ssh.Dial("tcp", cli.Address, &ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

// RunCmd runs commands
func (cli *SSHClient) RunCmd(cmd string) error {
	var buffer bytes.Buffer
	if cli.local {
		args := strings.Split(cmd, " ")
		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = &buffer
		if err := command.Run(); err != nil {
			return errors.New(buffer.String())
		}
	} else {
		session, err := cli.client.NewSession()
		if err != nil {
			return errors.Wrapf(err, "create ssh session failed, address=%s", cli.Address)
		}
		defer session.Close()
		session.Stdout = os.Stdout
		session.Stderr = &buffer
		if err = session.Run(cmd); err != nil {
			return errors.New(buffer.String())
		}
	}
	return nil
}

// Sftp transfers files
func (cli *SSHClient) Sftp(localFilePath string, remoteDir string) error {
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

// transferFile uses to transfer single file
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
		_, err := dstFile.Write(buf[:n])
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to write remote file")
		}
	}
	return nil
}

func (cli *SSHClient) Close() error {
	if cli.client != nil {
		return cli.client.Close()
	}
	return nil
}
