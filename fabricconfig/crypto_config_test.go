package fabricconfig

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"testing"
)

func TestUnmarshalCryptoConfig(t *testing.T) {
	data, err := ioutil.ReadFile("../sampleconfig/v1.4/crypto-config.yaml")
	require.NoError(t, err)
	cc := CryptoConfig{}
	err = yaml.Unmarshal(data, &cc)
	require.NoError(t, err)
	t.Log(cc)
}

func TestMarshalCryptoConfig(t *testing.T) {
	c := cryptoNodeConfig{
		Name:          "aa",
		Domain:        "example.com",
		EnableNodeOUs: true,
		Specs:         []cryptoSpec{{Hostname: "aa.example.com"}, {Hostname: "bb"}},
		Users:         cryptoUsers{Count: 1},
	}
	var cpc []cryptoNodeConfig
	cpc = append(cpc, c)
	cryptoConfig := CryptoConfig{
		PeerOrgs: cpc,
	}
	data, err := yaml.Marshal(cryptoConfig)
	require.NoError(t, err)
	err = ioutil.WriteFile("test.yaml", data, 0755)
	require.NoError(t, err)
}

func TestGenerateLocallyTestNetwork(t *testing.T) {
	err := GenerateLocallyTestNetworkCryptoConfig("./")
	require.NoError(t, err)
}

func TestSplitPeerOrOrdererUrl(t *testing.T) {
	urls := []string{"peer0.org1", "aa.org2.example.com"}

	for _, url := range urls {
		firstDotIndex := strings.Index(url, ".")
		peerName := url[:firstDotIndex]

		args := strings.Split(url, ".")
		orgName := args[1]
		domain := url[firstDotIndex+1:]

		t.Logf("peerName=%s, orgName=%s, domain=%s\n", peerName, orgName, domain)
	}
}

func TestGenerateCryptoConfigFile(t *testing.T) {
	err := GenerateCryptoConfigFile("./", nil, nil)
	require.NoError(t, err)
}
