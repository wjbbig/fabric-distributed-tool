package connectionprofile

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

func TestGenerateTestNetworkConnProfile(t *testing.T) {
	var connProfile ConnProfile
	connProfile.Version = "1.0.0"

	// client config
	client := &Client{
		Organization: "Org1",
		Logger:       &Logger{"info"},
		CryptoConfig: &CryptoConfig{"${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}"},
		CredentialStore: &CredentialStore{
			Path:        "/tmp/state-store",
			CryptoStore: &CryptoStore{"/tmp/msp"},
		},
		BCCSP: &BCCSP{&BCCSPSecurity{
			Enable:        true,
			Default:       &BCCSPDefault{"SW"},
			HashAlgorithm: "SHA2",
			SoftVerify:    true,
			Level:         256,
		}},
		TLSCerts: &TLSCerts{
			SystemCertPool: true,
			TLSClient: &TLSClient{
				Key:  &Key{"${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"},
				Cert: &Cert{"${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt"},
			},
		},
	}
	connProfile.Client = client

	// channels
	channels := map[string]*Channel{
		"mychannel": {
			Peers: map[string]*ChannelPeer{
				"peer0.org1.example.com": {
					EndorsingPeer:  true,
					ChaincodeQuery: true,
					LedgerQuery:    true,
					EventSource:    true,
				},
			},
		},
	}
	connProfile.Channels = channels

	// organizations config
	organizations := map[string]*Organization{
		"Org1": {
			MSPId:      "Org1MSP",
			CryptoPath: "peerOrganizations/org1.example.com/users/{username}@org1.example.com/msp",
			Peers: []string{
				"peer0.org1.example.com",
			},
			CertificateAuthorities: []string{
				"ca.org1.example.com",
			},
		},
	}
	connProfile.Organizations = organizations

	// orderer config
	orderers := map[string]*Orderer{
		"orderer.example.com": {
			URL: "orderer.example.com:7050",
			GRPCOptions: &GRPCOptions{
				SSLTargetNameOverride: "orderer.example.com",
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
				KeepAlivePermit:       false,
				FailFast:              false,
				AllowInsecure:         false,
			},
			TLSCACerts: &TLSCACert{Path: "${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"},
		},
	}
	connProfile.Orderers = orderers

	// peers config
	peers := map[string]*Peer{
		"peer0.org1.example.com": {
			URL: "peer0.org1.example.com:7051",
			GRPCOptions: &GRPCOptions{
				SSLTargetNameOverride: "peer0.org1.example.com",
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
				KeepAlivePermit:       false,
				FailFast:              false,
				AllowInsecure:         false,
			},
			TLSCACerts: &TLSCACert{Path: "${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"},
		},
	}
	connProfile.Peers = peers

	// ca config
	cas := map[string]*CertificateAuthority{
		"ca.org1.example.com": {
			URL: "https://ca.org1.example.com:7054",
			Registrar: &Registrar{
				EnrollId:     "admin",
				EnrollSecret: "adminpw",
			},
			TLSCACert: &TLSCACert{
				Path: "${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem",
				TLSClient: &TLSClient{
					Key:  &Key{"${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"},
					Cert: &Cert{"${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt"},
				},
			},
			CAName: "ca.org1.example.com",
		},
	}
	connProfile.CertificateAuthorities = cas

	data, err := yaml.Marshal(connProfile)
	require.NoError(t, err, "marshal connProfile failed")

	err = ioutil.WriteFile("./connection-config.yaml", data, 0755)
	require.NoError(t, err)
}

func TestGenerateNetworkConnProfile(t *testing.T) {
	GenerateNetworkConnProfile(".", "mychannel",
		[]string{"peer0.org1.example.com:7051:192.111.43.11", "peer0.org2.example.com:8051:127.0.0.1"},
		[]string{"orderer.example.com:7050:127.0.0.1"})
}
