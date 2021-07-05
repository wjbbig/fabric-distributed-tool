package fabricconfig

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

const defaultCryptoConfigFileName = "crypto-config.yaml"

type CryptoConfig struct {
	OrdererOrgs []cryptoOrdererConfig `yaml:"OrdererOrgs,omitempty"`
	PeerOrgs    []cryptoPeerConfig    `yaml:"PeerOrgs,omitempty"`
}

type cryptoOrdererConfig struct {
	Name          string       `yaml:"Name,omitempty"`
	Domain        string       `yaml:"Domain,omitempty"`
	EnableNodeOUs bool         `yaml:"EnableNodeOUs,omitempty"`
	Specs         []cryptoSpec `yaml:"Specs,omitempty"`
}

type cryptoPeerConfig struct {
	Name          string         `yaml:"Name,omitempty"`
	Domain        string         `yaml:"Domain,omitempty"`
	EnableNodeOUs bool           `yaml:"EnableNodeOUs,omitempty"`
	Specs         []cryptoSpec   `yaml:"Specs,omitempty"`
	Template      cryptoTemplate `yaml:"Template,omitempty"`
	Users         cryptoUsers    `yaml:"Users,omitempty"`
}

type cryptoSpec struct {
	Hostname   string `yaml:"Hostname,omitempty"`
	CommonName string `yaml:"CommonName,omitempty"`
}

type cryptoTemplate struct {
	Count    int    `yaml:"Count,omitempty"`
	Start    int    `yaml:"Start,omitempty"`
	Hostname string `yaml:"Hostname,omitempty"`
}

type cryptoUsers struct {
	Count int `yaml:"Count,omitempty"`
}

// GenerateCryptoConfigFile 根据传入的信息生成crypto-config.yaml文件
func GenerateCryptoConfigFile(filePath string) error {
	//path := filepath.Join(filePath, defaultCryptoConfigFileName)
	return nil
}

// GenerateLocallyTestNetwork 生成一个本地测试网络的crypto-config.yaml文件
func GenerateLocallyTestNetwork(filePath string) error {
	var cryptoConfig CryptoConfig
	var ordererConfigs []cryptoOrdererConfig
	var peerConfigs []cryptoPeerConfig
	ordererConfigs = append(ordererConfigs, cryptoOrdererConfig{
		Name:          "Orderer",
		Domain:        "example.com",
		EnableNodeOUs: true,
		Specs: []cryptoSpec{
			{
				Hostname: "orderer",
			},
			{
				Hostname: "orderer2",
			},
			{
				Hostname: "orderer3",
			},
			{
				Hostname: "orderer4",
			},
			{
				Hostname: "orderer5",
			},
		},
	})
	cryptoConfig.OrdererOrgs = ordererConfigs

	peerConfigs = append(peerConfigs, cryptoPeerConfig{
		Name:          "Org1",
		Domain:        "org1.example.com",
		EnableNodeOUs: true,
		Template:      cryptoTemplate{Count: 2},
		Users:         cryptoUsers{Count: 1},
	})
	peerConfigs = append(peerConfigs, cryptoPeerConfig{
		Name:          "Org2",
		Domain:        "org2.example.com",
		EnableNodeOUs: true,
		Template:      cryptoTemplate{Count: 2},
		Users:         cryptoUsers{Count: 1},
	})
	cryptoConfig.PeerOrgs = peerConfigs

	data, err := yaml.Marshal(cryptoConfig)
	if err != nil {
		return errors.Wrap(err, "yaml marshal failed")
	}

	filePath = filepath.Join(filePath, defaultCryptoConfigFileName)
	return ioutil.WriteFile(filePath, data, 0755)
}
