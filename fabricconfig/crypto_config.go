package fabricconfig

type CryptoConfig struct {
	OrdererOrgs struct {
		ordererConfigs []cryptoOrdererConfig
	} `yaml:"OrdererOrgs,omitempty"`
	PeerOrgs struct {
		peerConfigs []cryptoPeerConfig
	} `yaml:"PeerOrgs,omitempty"`
}

type cryptoOrdererConfig struct {
}

type cryptoPeerConfig struct {
	Name          string         `yaml:"Name,omitempty"`
	Domain        string         `yaml:"Domain,omitempty"`
	EnableNodeOUs bool           `yaml:"EnableNodeOUs,omitempty"`
	Template      cryptoTemplate `yaml:"Template,omitempty"`
	Users         cryptoUsers    `yaml:"Users,omitempty"`
}

// type cryptoSpec struct {
// }

type cryptoTemplate struct {
	Count int `yaml:"Count,omitempty"`
}

type cryptoUsers struct {
	Count int `yaml:"Count,omitempty"`
}

func GenerateCryptoConfigFile() error {

	return nil
}
