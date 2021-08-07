package fabricconfig

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/tools/cryptogen/ca"
	"github.com/wjbbig/fabric-distributed-tool/tools/cryptogen/msp"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

const defaultCryptoConfigFileName = "crypto-config.yaml"

var logger = log.NewLogger()

type CryptoConfig struct {
	OrdererOrgs []cryptoNodeConfig `yaml:"OrdererOrgs,omitempty"`
	PeerOrgs    []cryptoNodeConfig `yaml:"PeerOrgs,omitempty"`
}

type cryptoNodeConfig struct {
	Name          string         `yaml:"Name,omitempty"`
	Domain        string         `yaml:"Domain,omitempty"`
	EnableNodeOUs bool           `yaml:"EnableNodeOUs,omitempty"`
	CA            cryptoSpec     `yaml:"CA,omitempty"`
	Specs         []cryptoSpec   `yaml:"Specs,omitempty"`
	Template      cryptoTemplate `yaml:"Template,omitempty"`
	Users         cryptoUsers    `yaml:"Users,omitempty"`
}

type cryptoSpec struct {
	isAdmin            bool
	Hostname           string   `yaml:"Hostname,omitempty"`
	CommonName         string   `yaml:"CommonName,omitempty"`
	Country            string   `yaml:"Country,omitempty"`
	Province           string   `yaml:"Province,omitempty"`
	Locality           string   `yaml:"Locality,omitempty"`
	OrganizationalUnit string   `yaml:"Locality,omitempty"`
	StreetAddress      string   `yaml:"StreetAddress,omitempty"`
	PostalCode         string   `yaml:"PostalCode,omitempty"`
	SANS               []string `yaml:"SANS,omitempty"`
}

type cryptoTemplate struct {
	Count    int      `yaml:"Count,omitempty"`
	Start    int      `yaml:"Start,omitempty"`
	Hostname string   `yaml:"Hostname,omitempty"`
	SANS     []string `yaml:"SANS,omitempty"`
}

type cryptoUsers struct {
	Count int `yaml:"Count,omitempty"`
}

// GenerateCryptoConfigFile 根据传入的信息生成crypto-config.yaml文件
func GenerateCryptoConfigFile(filePath string, peers, orderers []string) error {
	logger.Info("begin to generate crypto-config.yaml")
	path := filepath.Join(filePath, defaultCryptoConfigFileName)
	var cryptoConfig CryptoConfig
	var ordererConfigs []cryptoNodeConfig
	var peerConfigs []cryptoNodeConfig

	ordererMap := make(map[string]cryptoNodeConfig)
	for _, ordererUrl := range orderers {
		ordererName, ordererOrg, ordererDomain := utils.SplitNameOrgDomain(ordererUrl)
		oc, ok := ordererMap[ordererDomain]
		if !ok {
			ordererMap[ordererDomain] = cryptoNodeConfig{
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
	peerMap := make(map[string]cryptoNodeConfig)
	for _, peerUrl := range peers {
		peerName, peerOrg, peerDomain := utils.SplitNameOrgDomain(peerUrl)
		pc, ok := peerMap[peerDomain]
		if !ok {
			peerMap[peerDomain] = cryptoNodeConfig{
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
	logger.Info("finish generating crypto-config.yaml")
	return ioutil.WriteFile(path, data, 0755)
}

// GenerateLocallyTestNetworkCryptoConfig 生成一个本地测试网络的crypto-config.yaml文件
func GenerateLocallyTestNetworkCryptoConfig(filePath string) error {
	var cryptoConfig CryptoConfig
	var ordererConfigs []cryptoNodeConfig
	var peerConfigs []cryptoNodeConfig
	ordererConfigs = append(ordererConfigs, cryptoNodeConfig{
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

	peerConfigs = append(peerConfigs, cryptoNodeConfig{
		Name:          "Org1",
		Domain:        "org1.example.com",
		EnableNodeOUs: true,
		Template:      cryptoTemplate{Count: 2},
		Users:         cryptoUsers{Count: 1},
	})
	peerConfigs = append(peerConfigs, cryptoNodeConfig{
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

// GenerateKeyPairsAndCerts 使用cryptogen工具生成peer和orderer的证书和密钥
func GenerateKeyPairsAndCerts(fileDir string) error {
	logger.Info("begin to generate key pairs and certs with crypto-config.yaml")
	defer logger.Info("finish generating key pairs and certs")

	var args []string
	args = append(args, "generate")
	args = append(args, fmt.Sprintf("--config=%s/%s", fileDir, defaultCryptoConfigFileName))
	args = append(args, fmt.Sprintf("--output=%s/%s", fileDir, "crypto-config"))
	cryptogenPath := filepath.Join("tools", "cryptogen")
	return utils.RunLocalCmd(cryptogenPath, args...)
}

func renderOrgSpec(nodeConfig *cryptoNodeConfig, prefix string) error {
	return nil
}

func generatePeerOrgKeypairsAndCerts(baseDir string, nodeConfig cryptoNodeConfig) error {
	orgName := nodeConfig.Domain
	// generate CAs
	orgDir := filepath.Join(baseDir, "peerOrganizations", orgName)
	caDir := filepath.Join(orgDir, "ca")
	tlsCADir := filepath.Join(orgDir, "tlsca")
	mspDir := filepath.Join(orgDir, "msp")
	peersDir := filepath.Join(orgDir, "peers")
	usersDir := filepath.Join(orgDir, "users")
	adminCertsDir := filepath.Join(mspDir, "admincerts")

	signCA, err := ca.NewCA(caDir, orgName, nodeConfig.CA.CommonName, nodeConfig.CA.Country, nodeConfig.CA.Province,
		nodeConfig.CA.Locality, nodeConfig.CA.OrganizationalUnit, nodeConfig.CA.StreetAddress, nodeConfig.CA.PostalCode)
	if err != nil {
		return errors.Wrapf(err, "failed to generate signCA for org %s", orgName)
	}
	tlsCA, err := ca.NewCA(tlsCADir, orgName, "tls"+nodeConfig.CA.CommonName, nodeConfig.CA.Country, nodeConfig.CA.Province,
		nodeConfig.CA.Locality, nodeConfig.CA.OrganizationalUnit, nodeConfig.CA.StreetAddress, nodeConfig.CA.PostalCode)
	if err != nil {
		return errors.Wrapf(err, "failed to genreating MSP for org %s", orgName)
	}
	// generate TLS CA
	if err := msp.GenerateVerifyingMSP(mspDir, signCA, tlsCA, nodeConfig.EnableNodeOUs); err != nil {
		return errors.Wrapf(err, "failed to generate MSP for org %s", orgName)
	}

	return nil
}
