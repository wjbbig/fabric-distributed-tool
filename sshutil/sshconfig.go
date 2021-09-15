package sshutil

import (
	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const defaultSSHConfigName = "ssh-config.yaml"

type SSHConfig struct {
	Clients []Client `yaml:"clients,omitempty"`
}

type Client struct {
	Name     string `yaml:"name,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
}

// GenerateSSHConfig stores node info
// clientUrl example: peer0.org1.example.com:7050@root@127.0.0.1:22:password
func GenerateSSHConfig(filePath string, clients []Client) error {
	var sshConf SSHConfig
	sshConf.Clients = clients
	filePath = filepath.Join(filePath, defaultSSHConfigName)
	data, err := yaml.Marshal(sshConf)
	if err != nil {
		return errors.Wrap(err, "failed to marshal sshConfig")
	}
	return ioutil.WriteFile(filePath, data, 0755)
}

func NewClient(clientUrl string, nodeType string) Client {
	args := strings.Split(clientUrl, "@")
	name := strings.Split(args[0], ":")[0]
	username := args[1]
	if len(args) > 3 {
		for i := 3; i < len(args); i++ {
			args[2] += "@"
			args[2] += args[i]
		}
	}
	indexes := utils.Indexes(args[2], ":")
	return Client{
		Name:     name,
		Username: username,
		Type:     nodeType,
		Password: args[2][indexes[1]+1:],
		Host:     args[2][:indexes[0]],
		Port:     args[2][indexes[0]+1 : indexes[1]],
	}
}

func UnmarshalSSHConfig(filePath string) (*SSHConfig, error) {
	sshConf := &SSHConfig{}
	filePath = filepath.Join(filePath, defaultSSHConfigName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, sshConf); err != nil {
		return nil, err
	}

	return sshConf, nil
}
