package fabricconfig

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"github.com/wjbbig/fabric-distributed-tool/tools/cryptogen/ca"
	"github.com/wjbbig/fabric-distributed-tool/tools/cryptogen/csp"
	"github.com/wjbbig/fabric-distributed-tool/tools/cryptogen/msp"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const (
	defaultCryptoConfigFileName = "crypto-config.yaml"
	userBaseName                = "User"
	adminBaseName               = "Admin"
	defaultHostnameTemplate     = "{{.Prefix}}{{.Index}}"
	defaultCNTemplate           = "{{.Hostname}}.{{.Domain}}"
)

type HostnameData struct {
	Prefix string
	Index  int
	Domain string
}

type SpecData struct {
	Hostname   string
	Domain     string
	CommonName string
}

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
	OrganizationalUnit string   `yaml:"OrganizationalUnit,omitempty"`
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
func GenerateCryptoConfigFile(filePath string, peers, orderers []*network.Node) error {
	logger.Info("begin to generate crypto-config.yaml")
	path := filepath.Join(filePath, defaultCryptoConfigFileName)
	var cryptoConfig CryptoConfig
	var ordererConfigs []cryptoNodeConfig
	var peerConfigs []cryptoNodeConfig

	ordererMap := make(map[string]cryptoNodeConfig)
	for _, node := range orderers {
		oc, ok := ordererMap[node.Domain]
		if !ok {
			ordererMap[node.Domain] = cryptoNodeConfig{
				Name:          node.OrgId,
				Domain:        node.Domain,
				EnableNodeOUs: true,
				Specs:         []cryptoSpec{{Hostname: node.Name}},
			}
		} else {
			oc.Specs = append(oc.Specs, cryptoSpec{Hostname: node.Name})
			ordererMap[node.Domain] = oc
		}

	}
	for _, config := range ordererMap {
		ordererConfigs = append(ordererConfigs, config)
	}
	peerMap := make(map[string]cryptoNodeConfig)
	for _, node := range peers {
		pc, ok := peerMap[node.Domain]
		if !ok {
			peerMap[node.Domain] = cryptoNodeConfig{
				Name:          node.OrgId,
				Domain:        node.Domain,
				EnableNodeOUs: true,
				Specs:         []cryptoSpec{{Hostname: node.Name}},
				Users:         cryptoUsers{Count: 1},
			}
		} else {
			pc.Specs = append(pc.Specs, cryptoSpec{Hostname: node.Name})
			peerMap[node.Domain] = pc
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
	if err := ioutil.WriteFile(path, data, 0755); err != nil {
		return err
	}
	logger.Info("finish generating crypto-config.yaml")
	return nil
}

func ExtendCryptoConfigFile(filePath string, peers, orderers []string) error {
	logger.Info("begin to extend crypto-config.yaml")
	path := filepath.Join(filePath, defaultCryptoConfigFileName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "crypto-config.yaml not found")
	}
	var cryptoConfig CryptoConfig
	if err := yaml.Unmarshal(data, &cryptoConfig); err != nil {
		return errors.Wrap(err, "failed to unmarshal cryptoConfig")
	}

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
		cryptoConfig.OrdererOrgs = append(cryptoConfig.OrdererOrgs, config)
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
		cryptoConfig.PeerOrgs = append(cryptoConfig.PeerOrgs, config)
	}

	data, err = yaml.Marshal(cryptoConfig)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, data, 0755); err != nil {
		return err
	}
	logger.Info("finish extending crypto-config.yaml")
	return nil
}

// GenerateLocallyTestNetworkCryptoConfig generates a crypto config file for testing
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

// GenerateKeyPairsAndCerts generates key pairs and certs for peer and orderer
func GenerateKeyPairsAndCerts(fileDir string) error {
	logger.Info("begin to generate key pairs and certs with crypto-config.yaml")

	data, err := ioutil.ReadFile(filepath.Join(fileDir, defaultCryptoConfigFileName))
	if err != nil {
		return err
	}
	var config CryptoConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return errors.Wrapf(err, "failed to unmarshal yaml data")
	}
	outputDir := filepath.Join(fileDir, "crypto-config")
	for _, org := range config.PeerOrgs {
		if err := renderOrgSpec(&org, "peer"); err != nil {
			return err
		}
		if err := generatePeerOrg(outputDir, org); err != nil {
			return err
		}
	}

	for _, org := range config.OrdererOrgs {
		if err := renderOrgSpec(&org, "orderer"); err != nil {
			return err
		}
		if err := generateOrdererOrg(outputDir, org); err != nil {
			return err
		}
	}
	logger.Info("finish generating key pairs and certs")
	return nil
}

// ExtendKeyPairsAndCerts extends key pairs and certs for peer and orderer
func ExtendKeyPairsAndCerts(fileDir string) error {
	logger.Info("begin to extend key pairs and certs with crypto-config.yaml")

	data, err := ioutil.ReadFile(filepath.Join(fileDir, defaultCryptoConfigFileName))
	if err != nil {
		return err
	}
	var config CryptoConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return errors.Wrapf(err, "failed to unmarshal yaml data")
	}
	outputDir := filepath.Join(fileDir, "crypto-config")
	for _, org := range config.PeerOrgs {
		if err := renderOrgSpec(&org, "peer"); err != nil {
			return err
		}
		if err := extendPeerOrg(outputDir, org); err != nil {
			return err
		}
	}

	for _, org := range config.OrdererOrgs {
		if err := renderOrgSpec(&org, "orderer"); err != nil {
			return err
		}
		if err := extendOrdererOrg(outputDir, org); err != nil {
			return err
		}
	}
	logger.Info("finish extending key pairs and certs with crypto-config.yaml")
	return nil
}

func renderOrgSpec(nodeConfig *cryptoNodeConfig, prefix string) error {
	for i := 0; i < nodeConfig.Template.Count; i++ {
		data := HostnameData{
			Prefix: prefix,
			Index:  i + nodeConfig.Template.Start,
			Domain: nodeConfig.Domain,
		}

		hostname, err := parseTemplateWithDefault(nodeConfig.Template.Hostname, defaultHostnameTemplate, data)
		if err != nil {
			return err
		}

		spec := cryptoSpec{
			Hostname: hostname,
			SANS:     nodeConfig.Template.SANS,
		}
		nodeConfig.Specs = append(nodeConfig.Specs, spec)
	}

	// Touch up all general node-specs to add the domain
	for idx, spec := range nodeConfig.Specs {
		err := renderNodeSpec(nodeConfig.Domain, &spec)
		if err != nil {
			return err
		}

		nodeConfig.Specs[idx] = spec
	}

	// Process the CA node-spec in the same manner
	if len(nodeConfig.CA.Hostname) == 0 {
		nodeConfig.CA.Hostname = "ca"
	}
	err := renderNodeSpec(nodeConfig.Domain, &nodeConfig.CA)
	if err != nil {
		return err
	}

	return nil
}

func generatePeerOrg(baseDir string, nodeConfig cryptoNodeConfig) error {
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
	if err := generateNodes(peersDir, nodeConfig.Specs, signCA, tlsCA, msp.PEER, nodeConfig.EnableNodeOUs); err != nil {
		return err
	}

	var users []cryptoSpec
	for i := 1; i <= nodeConfig.Users.Count; i++ {
		user := cryptoSpec{
			CommonName: fmt.Sprintf("%s%d@%s", userBaseName, i, orgName),
		}

		users = append(users, user)
	}

	adminUser := cryptoSpec{
		isAdmin:    true,
		CommonName: fmt.Sprintf("%s@%s", adminBaseName, orgName),
	}
	users = append(users, adminUser)
	if err := generateNodes(usersDir, users, signCA, tlsCA, msp.CLIENT, nodeConfig.EnableNodeOUs); err != nil {
		return err
	}

	// copy the admin cert to the org's MSP admincerts
	if !nodeConfig.EnableNodeOUs {
		err = copyAdminCert(usersDir, adminCertsDir, adminUser.CommonName)
		if err != nil {
			fmt.Printf("Error copying admin cert for org %s:\n%v\n",
				orgName, err)
			os.Exit(1)
		}
	}

	// copy the admin cert to each of the org's peer's MSP admincerts
	for _, spec := range nodeConfig.Specs {
		if nodeConfig.EnableNodeOUs {
			continue
		}
		err = copyAdminCert(usersDir,
			filepath.Join(peersDir, spec.CommonName, "msp", "admincerts"), adminUser.CommonName)
		if err != nil {
			return errors.Wrapf(err, "failed to copying admin cert for org %s peer %s", orgName, spec.CommonName)
		}
	}
	return nil
}

func generateOrdererOrg(baseDir string, orgSpec cryptoNodeConfig) error {
	orgName := orgSpec.Domain
	orgDir := filepath.Join(baseDir, "ordererOrganizations", orgName)
	caDir := filepath.Join(orgDir, "ca")
	tlsCADir := filepath.Join(orgDir, "tlsca")
	mspDir := filepath.Join(orgDir, "msp")
	orderersDir := filepath.Join(orgDir, "orderers")
	usersDir := filepath.Join(orgDir, "users")
	adminCertsDir := filepath.Join(mspDir, "admincerts")
	// generate signing CA
	signCA, err := ca.NewCA(caDir, orgName, orgSpec.CA.CommonName, orgSpec.CA.Country, orgSpec.CA.Province, orgSpec.CA.Locality, orgSpec.CA.OrganizationalUnit, orgSpec.CA.StreetAddress, orgSpec.CA.PostalCode)
	if err != nil {
		return errors.Wrapf(err, "failed to generating signCA for org %s", orgName)
	}
	// generate TLS CA
	tlsCA, err := ca.NewCA(tlsCADir, orgName, "tls"+orgSpec.CA.CommonName, orgSpec.CA.Country, orgSpec.CA.Province, orgSpec.CA.Locality, orgSpec.CA.OrganizationalUnit, orgSpec.CA.StreetAddress, orgSpec.CA.PostalCode)
	if err != nil {
		return errors.Wrapf(err, "failed to generating tlsCA for org %s", orgName)
	}

	err = msp.GenerateVerifyingMSP(mspDir, signCA, tlsCA, orgSpec.EnableNodeOUs)
	if err != nil {
		return errors.Wrapf(err, "failed to generating MSP for org %s", orgName)
	}

	if err := generateNodes(orderersDir, orgSpec.Specs, signCA, tlsCA, msp.ORDERER, orgSpec.EnableNodeOUs); err != nil {
		return err
	}

	adminUser := cryptoSpec{
		isAdmin:    true,
		CommonName: fmt.Sprintf("%s@%s", adminBaseName, orgName),
	}

	// generate an admin for the orderer org
	users := []cryptoSpec{}
	// add an admin user
	users = append(users, adminUser)
	if err := generateNodes(usersDir, users, signCA, tlsCA, msp.CLIENT, orgSpec.EnableNodeOUs); err != nil {
		return err
	}

	// copy the admin cert to the org's MSP admincerts
	if !orgSpec.EnableNodeOUs {
		err = copyAdminCert(usersDir, adminCertsDir, adminUser.CommonName)
		if err != nil {
			return errors.Wrapf(err, "failed to copying admin cert for org %s", orgName)
		}
	}

	// copy the admin cert to each of the org's orderers's MSP admincerts
	for _, spec := range orgSpec.Specs {
		if orgSpec.EnableNodeOUs {
			continue
		}
		err = copyAdminCert(usersDir,
			filepath.Join(orderersDir, spec.CommonName, "msp", "admincerts"), adminUser.CommonName)
		if err != nil {
			return errors.Wrapf(err, "failed to copying admin cert for org %s orderer %s", orgName, spec.CommonName)
		}
	}
	return nil
}

func extendPeerOrg(fileDir string, orgSpec cryptoNodeConfig) error {
	orgName := orgSpec.Domain
	orgDir := filepath.Join(fileDir, "peerOrganizations", orgName)
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		if err := generatePeerOrg(fileDir, orgSpec); err != nil {
			return err
		}
	}

	peersDir := filepath.Join(orgDir, "peers")
	usersDir := filepath.Join(orgDir, "users")
	caDir := filepath.Join(orgDir, "ca")
	tlscaDir := filepath.Join(orgDir, "tlsca")

	signCA := getCA(caDir, orgSpec, orgSpec.CA.CommonName)
	tlsCA := getCA(tlscaDir, orgSpec, "tls"+orgSpec.CA.CommonName)

	if err := generateNodes(peersDir, orgSpec.Specs, signCA, tlsCA, msp.PEER, orgSpec.EnableNodeOUs); err != nil {
		return err
	}

	adminUser := cryptoSpec{
		isAdmin:    true,
		CommonName: fmt.Sprintf("%s@%s", adminBaseName, orgName),
	}
	// copy the admin cert to each of the org's peer's MSP admincerts
	for _, spec := range orgSpec.Specs {
		if orgSpec.EnableNodeOUs {
			continue
		}
		err := copyAdminCert(usersDir,
			filepath.Join(peersDir, spec.CommonName, "msp", "admincerts"), adminUser.CommonName)
		if err != nil {
			return errors.Wrapf(err, "failed to copying admin cert for org %s peer %s", orgName, spec.CommonName)
		}
	}

	var users []cryptoSpec
	for j := 1; j <= orgSpec.Users.Count; j++ {
		user := cryptoSpec{
			CommonName: fmt.Sprintf("%s%d@%s", userBaseName, j, orgName),
		}

		users = append(users, user)
	}

	if err := generateNodes(usersDir, users, signCA, tlsCA, msp.CLIENT, orgSpec.EnableNodeOUs); err != nil {
		return err
	}

	return nil
}
func extendOrdererOrg(fileDir string, orgSpec cryptoNodeConfig) error {
	orgName := orgSpec.Domain

	orgDir := filepath.Join(fileDir, "ordererOrganizations", orgName)
	caDir := filepath.Join(orgDir, "ca")
	usersDir := filepath.Join(orgDir, "users")
	tlscaDir := filepath.Join(orgDir, "tlsca")
	orderersDir := filepath.Join(orgDir, "orderers")
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		if err := generatePeerOrg(fileDir, orgSpec); err != nil {
			return err
		}
	}

	signCA := getCA(caDir, orgSpec, orgSpec.CA.CommonName)
	tlsCA := getCA(tlscaDir, orgSpec, "tls"+orgSpec.CA.CommonName)

	if err := generateNodes(orderersDir, orgSpec.Specs, signCA, tlsCA, msp.ORDERER, orgSpec.EnableNodeOUs); err != nil {
		return err
	}

	adminUser := cryptoSpec{
		isAdmin:    true,
		CommonName: fmt.Sprintf("%s@%s", adminBaseName, orgName),
	}

	for _, spec := range orgSpec.Specs {
		if orgSpec.EnableNodeOUs {
			continue
		}
		err := copyAdminCert(usersDir,
			filepath.Join(orderersDir, spec.CommonName, "msp", "admincerts"), adminUser.CommonName)
		if err != nil {
			return errors.Wrapf(err, "failed to copying admin cert for org %s peer %s", orgName, spec.CommonName)
		}
	}

	return nil
}

func copyAdminCert(usersDir, adminCertsDir, adminUserName string) error {
	if _, err := os.Stat(filepath.Join(adminCertsDir,
		adminUserName+"-cert.pem")); err == nil {
		return nil
	}
	// delete the contents of admincerts
	err := os.RemoveAll(adminCertsDir)
	if err != nil {
		return err
	}
	// recreate the admincerts directory
	err = os.MkdirAll(adminCertsDir, 0755)
	if err != nil {
		return err
	}
	err = copyFile(filepath.Join(usersDir, adminUserName, "msp", "signcerts",
		adminUserName+"-cert.pem"), filepath.Join(adminCertsDir,
		adminUserName+"-cert.pem"))
	if err != nil {
		return err
	}
	return nil
}

func generateNodes(baseDir string, nodes []cryptoSpec, signCA *ca.CA, tlsCA *ca.CA, nodeType int, nodeOUs bool) error {
	for _, node := range nodes {
		nodeDir := filepath.Join(baseDir, node.CommonName)
		if _, err := os.Stat(nodeDir); os.IsNotExist(err) {
			currentNodeType := nodeType
			if node.isAdmin && nodeOUs {
				currentNodeType = msp.ADMIN
			}
			err := msp.GenerateLocalMSP(nodeDir, node.CommonName, node.SANS, signCA, tlsCA, currentNodeType, nodeOUs)
			if err != nil {
				return errors.Wrapf(err, "failed to generating local MSP for %v", node)
			}
		}
	}
	return nil
}

func parseTemplateWithDefault(input, defaultInput string, data interface{}) (string, error) {
	// Use the default if the input is an empty string
	if len(input) == 0 {
		input = defaultInput
	}

	return parseTemplate(input, data)
}

func parseTemplate(input string, data interface{}) (string, error) {

	t, err := template.New("parse").Parse(input)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %s", err)
	}

	output := new(bytes.Buffer)
	err = t.Execute(output, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %s", err)
	}

	return output.String(), nil
}

func renderNodeSpec(domain string, spec *cryptoSpec) error {
	data := SpecData{
		Hostname: spec.Hostname,
		Domain:   domain,
	}

	// Process our CommonName
	cn, err := parseTemplateWithDefault(spec.CommonName, defaultCNTemplate, data)
	if err != nil {
		return err
	}

	spec.CommonName = cn
	data.CommonName = cn

	// Save off our original, unprocessed SANS entries
	origSANS := spec.SANS

	// Set our implicit SANS entries for CN/Hostname
	spec.SANS = []string{cn, spec.Hostname}

	// Finally, process any remaining SANS entries
	for _, _san := range origSANS {
		san, err := parseTemplate(_san, data)
		if err != nil {
			return err
		}

		spec.SANS = append(spec.SANS, san)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func getCA(caDir string, spec cryptoNodeConfig, name string) *ca.CA {
	priv, _ := csp.LoadPrivateKey(caDir)
	cert, _ := ca.LoadCertificateECDSA(caDir)

	return &ca.CA{
		Name:               name,
		Signer:             priv,
		SignCert:           cert,
		Country:            spec.CA.Country,
		Province:           spec.CA.Province,
		Locality:           spec.CA.Locality,
		OrganizationalUnit: spec.CA.OrganizationalUnit,
		StreetAddress:      spec.CA.StreetAddress,
		PostalCode:         spec.CA.PostalCode,
	}
}
