package fabricconfig

import (
	"github.com/pkg/errors"
	log "github.com/wjbbig/fabric-distributed-tool/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const defaultCryptoConfigFileName = "crypto-config.yaml"

var logger = log.NewLogger(log.Info)

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
func GenerateCryptoConfigFile(filePath string, peers, orderers []string) error {
	logger.Info("begin to generate crypto-config.yaml")
	path := filepath.Join(filePath, defaultCryptoConfigFileName)
	var cryptoConfig CryptoConfig
	var ordererConfigs []cryptoOrdererConfig
	var peerConfigs []cryptoPeerConfig

	ordererMap := make(map[string]cryptoOrdererConfig)
	for _, ordererUrl := range orderers {
		ordererName, ordererOrg, ordererDomain := splitNameOrgDomain(ordererUrl)
		oc, ok := ordererMap[ordererDomain]
		if !ok {
			ordererMap[ordererDomain] = cryptoOrdererConfig{
				Name:          ordererOrg,
				Domain:        ordererDomain,
				EnableNodeOUs: true,
				Specs:         []cryptoSpec{{Hostname: ordererName}},
			}
		} else {
			oc.Specs = append(oc.Specs, cryptoSpec{Hostname: ordererName})
			ordererMap[ordererDomain] = oc
		}

	}
	for _, config := range ordererMap {
		ordererConfigs = append(ordererConfigs, config)
	}
	peerMap := make(map[string]cryptoPeerConfig)
	for _, peerUrl := range peers {
		peerName, peerOrg, peerDomain := splitNameOrgDomain(peerUrl)
		pc, ok := peerMap[peerDomain]
		if !ok {
			peerMap[peerDomain] = cryptoPeerConfig{
				Name:          peerOrg,
				Domain:        peerDomain,
				EnableNodeOUs: true,
				Specs:         []cryptoSpec{{Hostname: peerName}},
				Users:         cryptoUsers{Count: 1},
			}
		} else {
			pc.Specs = append(pc.Specs, cryptoSpec{Hostname: peerName})
			peerMap[peerDomain] = pc
		}
	}
	for _, config := range peerMap {
		peerConfigs = append(peerConfigs, config)
	}

	cryptoConfig.PeerOrgs = peerConfigs
	cryptoConfig.OrdererOrgs = ordererConfigs
	data, err := yaml.Marshal(cryptoConfig)
	if err != nil {
		return err
	}
	defer logger.Debugf("finish generating crypto-config.yaml, the content: \n%s", string(data))
	return ioutil.WriteFile(path, data, 0755)
}

// SplitNameOrgDomain 将url拆分成节点名称,组织名称和域名
// 默认以'.'为分割符,分割后第1个元素是节点名称,第二个是组织名,
// 第二个到之后所有的内容组为域名
func splitNameOrgDomain(url string) (string, string, string) {
	firstDotIndex := strings.Index(url, ".")
	name := url[:firstDotIndex]

	args := strings.Split(url, ".")
	orgName := args[1]
	domain := url[firstDotIndex+1:]
	return name, orgName, domain
}

// GenerateLocallyTestNetworkCryptoConfig 生成一个本地测试网络的crypto-config.yaml文件
func GenerateLocallyTestNetworkCryptoConfig(filePath string) error {
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
