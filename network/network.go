package network

type NetworkConfig struct {
	Name      string
	Channels  map[string]*Channel
	Nodes     map[string]*Node
	Chaincode map[string]*Chaincode
}

type Node struct {
	Name string
	Port string
	Type string
}

type Channel struct {
	Nodes     []string
	Chaincode []string
}

type Chaincode struct {
	Name      string
	Path      string
	Version   string
	InitParam string
}
