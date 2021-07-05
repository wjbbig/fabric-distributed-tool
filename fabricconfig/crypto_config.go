package fabricconfig

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

func GenerateCryptoConfigFile(filePath string) error {
	//path := filepath.Join(filePath, defaultCryptoConfigFileName)
	return nil
}
