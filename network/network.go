package network

type NetworkConfig struct {
	Name       string                `yaml:"name,omitempty"`
	Channels   map[string]*Channel   `yaml:"channels,omitempty"`
	Nodes      map[string]*Node      `yaml:"nodes,omitempty"`
	Chaincodes map[string]*Chaincode `yaml:"chaincodes,omitempty"`
}

type Node struct {
	Name     string `yaml:"name,omitempty"`
	NodePort int    `json:"node_port,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Host     string `yaml:"host,omitempty"`
	SSHPort  int    `yaml:"ssh_port,omitempty"`
	Couch    bool   `yaml:"couch,omitempty"`
}

type Channel struct {
	Nodes      []string `yaml:"nodes,omitempty"`
	Chaincodes []string `yaml:"chaincodes,omitempty"`
}

type Chaincode struct {
	Path      string `yaml:"path,omitempty"`
	Version   string `yaml:"version,omitempty"`
	InitParam string `yaml:"init_param,omitempty"`
}

func GenerateNetworkConfig(fileDir, networkName, channelId, ccId, ccPath, ccVersion, ccInitParam string, peerUrls, ordererUrls []string) error {
	var network NetworkConfig

	network.Name = networkName
	chaincode := &Chaincode{
		Path:      ccPath,
		Version:   ccVersion,
		InitParam: ccInitParam,
	}
	network.Chaincodes = map[string]*Chaincode{
		ccId: chaincode,
	}
	return nil
}
