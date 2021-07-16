package connectionprofile

import (
	"fmt"
	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	defaultConnProfileName     = "connection-config.yaml"
	defaultCryptoConfigDirName = "crypto-config"
)

var logger = mylogger.NewLogger()

// ConnProfile 生成fabric-sdk-go使用的连接文件
type ConnProfile struct {
	Version                string                          `yaml:"version,omitempty"`
	Client                 Client                          `yaml:"client,omitempty"`
	Channels               map[string]Channel              `yaml:"channels,omitempty"`
	Organizations          map[string]Organization         `yaml:"organizations,omitempty"`
	Orderers               map[string]Orderer              `yaml:"orderers,omitempty"`
	Peers                  map[string]Peer                 `yaml:"peers,omitempty"`
	CertificateAuthorities map[string]CertificateAuthority `yaml:"certificateAuthorities,omitempty"`
	EntityMatchers         map[string][]EntityMatcher      `yaml:"entityMatchers,omitempty"`
}

// ============================CLIENT==================================
type Client struct {
	Organization    string          `yaml:"organization,omitempty"`
	Logger          Logger          `yaml:"logger,omitempty"`
	CryptoConfig    CryptoConfig    `yaml:"cryptoconfig,omitempty"`
	CredentialStore CredentialStore `yaml:"credentialStore,omitempty"`
	BCCSP           BCCSP           `yaml:"BCCSP,omitempty"`
	TLSCerts        TLSCerts        `yaml:"tlsCerts,omitempty"`
}

type Logger struct {
	Level string `yaml:"level,omitempty"`
}

type CryptoConfig struct {
	Path string `yaml:"path,omitempty"`
}

type CredentialStore struct {
	Path        string      `yaml:"path,omitempty"`
	CryptoStore CryptoStore `yaml:"cryptoStore,omitempty"`
}

type CryptoStore struct {
	Path string `yaml:"path,omitempty"`
}

type BCCSP struct {
	Security BCCSPSecurity `yaml:"security,omitempty"`
}

type BCCSPSecurity struct {
	Enable        bool         `yaml:"enable,omitempty"`
	Default       BCCSPDefault `yaml:"default,omitempty"`
	HashAlgorithm string       `yaml:"hashAlgorithm,omitempty"`
	SoftVerify    bool         `yaml:"softVerify,omitempty"`
	Level         int          `yaml:"level,omitempty"`
}

type BCCSPDefault struct {
	Provider string `yaml:"provider,omitempty"`
}

type TLSCerts struct {
	SystemCertPool bool      `yaml:"systemCertPool,omitempty"`
	TLSClient      TLSClient `yaml:"client,omitempty"`
}

type TLSClient struct {
	Key  Key  `yaml:"key,omitempty"`
	Cert Cert `yaml:"cert,omitempty"`
}

type Key struct {
	Path string `yaml:"path,omitempty"`
}

type Cert struct {
	Path string `yaml:"path,omitempty"`
}

// ========================CHANNEL===========================
type Channel struct {
	Peers    map[string]ChannelPeer `yaml:"peers,omitempty"`
	Policies ChannelPolicies        `yaml:"policies,omitempty"`
}

type ChannelPeer struct {
	EndorsingPeer  bool `yaml:"endorsingPeer,omitempty"`
	ChaincodeQuery bool `yaml:"chaincodeQuery,omitempty"`
	LedgerQuery    bool `yaml:"ledgerQuery,omitempty"`
	EventSource    bool `yaml:"eventSource,omitempty"`
}

type ChannelPolicies struct {
	QueryChannelConfig QueryChannelConfig `yaml:"queryChannelConfig,omitempty"`
	Discovery          Discovery          `yaml:"discovery,omitempty"`
	EventService       EventService       `yaml:"eventService,omitempty"`
}

type QueryChannelConfig struct {
	MinResponse int `yaml:"minResponse,omitempty"`
	MaxTargets  int `yaml:"maxTargets,omitempty"`
	RetryOpts   struct {
		Attempts       int    `yaml:"attempts,omitempty"`
		InitialBackoff string `yaml:"initialBackoff,omitempty"`
		MaxBackoff     int    `yaml:"maxBackoff,omitempty"`
		BackoffFactor  string `yaml:"backoffFactor,omitempty"`
	} `yaml:"retryOpts,omitempty"`
}

type Discovery struct {
	MaxTargets int `yaml:"maxTargets,omitempty"`
	RetryOpts  struct {
		Attempts       int    `yaml:"attempts,omitempty"`
		InitialBackoff string `yaml:"initialBackoff,omitempty"`
		MaxBackoff     int    `yaml:"maxBackoff,omitempty"`
		BackoffFactor  string `yaml:"backoffFactor,omitempty"`
	} `yaml:"retryOpts,omitempty"`
}

type EventService struct {
	ResolverStrategy                 string `yaml:"resolverStrategy,omitempty"`
	Balancer                         string `yaml:"balancer,omitempty"`
	BlockHeightLagThreshold          int    `yaml:"blockHeightLagThreshold,omitempty"`
	ReconnectBlockHeightLagThreshold int    `yaml:"reconnectBlockHeightLagThreshold,omitempty"`
	PeerMonitorPeriod                string `yaml:"peerMonitorPeriod,omitempty"`
}

// ==========================ORGANIZATION=================================
type Organization struct {
	MSPId                  string   `yaml:"mspid,omitempty"`
	CryptoPath             string   `yaml:"cryptoPath,omitempty"`
	Peers                  []string `yaml:"peers,omitempty"`
	CertificateAuthorities []string `yaml:"certificateAuthorities,omitempty"`
}

// ==============================ORDERER==================================
type Orderer struct {
	URL         string      `yaml:"url,omitempty"`
	GRPCOptions GRPCOptions `yaml:"grpcOptions,omitempty"`
	TLSCACerts  TLSCACert   `yaml:"tlsCACerts,omitempty"`
}

// ==============================PEER=====================================
type Peer struct {
	URL         string      `yaml:"url,omitempty"`
	GRPCOptions GRPCOptions `yaml:"grpcOptions,omitempty"`
	TLSCACerts  TLSCACert   `yaml:"tlsCACerts,omitempty"`
}

type GRPCOptions struct {
	SSLTargetNameOverride string `yaml:"ssl-target-name-override,omitempty"`
	KeepAliveTime         string `yaml:"keep-alive-time,omitempty"`
	KeepAliveTimeout      string `yaml:"keep-alive-timeout,omitempty"`
	KeepAlivePermit       bool   `yaml:"keep-alive-permit,omitempty"`
	FailFast              bool   `yaml:"fail-fast,omitempty"`
	AllowInsecure         bool   `yaml:"allow-insecure,omitempty"`
}

// ============================CA=================================
type TLSCACert struct {
	Path      string    `yaml:"path,omitempty"`
	TLSClient TLSClient `yaml:"client,omitempty"`
}

type Registrar struct {
	EnrollId     string `yaml:"enrollId,omitempty"`
	EnrollSecret string `yaml:"enrollSecret,omitempty"`
}

type CertificateAuthority struct {
	URL       string    `yaml:"url,omitempty"`
	Registrar Registrar `yaml:"registrar,omitempty"`
	TLSCACert TLSCACert `yaml:"tlsCACerts,omitempty"`
	CAName    string    `yaml:"caName,omitempty"`
}

// =============================ENTITYMATCHER================================
type EntityMatcher struct {
	Pattern                             string `yaml:"pattern,omitempty"`
	UrlSubstitutionExp                  string `yaml:"urlSubstitutionExp,omitempty"`
	SSLTargetOverrideUrlSubstitutionExp string `yaml:"sslTargetOverrideUrlSubstitutionExp,omitempty"`
	MappedHost                          string `yaml:"mappedHost,omitempty"`
}

// GenerateNetworkConnProfile 生成连接文件,peer和orderer的格式必须是url:port:ip
func GenerateNetworkConnProfile(filePath string, channelId string, peerUrls, ordererUrls []string) error {
	logger.Info("begin to generate fabric-sdk-go connection profile")
	var connProfile ConnProfile
	connProfile.Version = "1.0.0"
	client := Client{
		Logger:       Logger{"info"},
		CryptoConfig: CryptoConfig{filepath.Join(filePath, defaultCryptoConfigDirName)},
		CredentialStore: CredentialStore{
			Path:        "/tmp/state-store",
			CryptoStore: CryptoStore{"/tmp/msp"},
		},
		BCCSP: BCCSP{BCCSPSecurity{
			Enable:        true,
			Default:       BCCSPDefault{"SW"},
			HashAlgorithm: "SHA2",
			SoftVerify:    true,
			Level:         256,
		}},
		TLSCerts: TLSCerts{
			SystemCertPool: false,
		},
	}
	connProfile.Client = client

	channel := Channel{Peers: map[string]ChannelPeer{}}
	peers := make(map[string]Peer)
	organizations := make(map[string]Organization)
	entityMatchers := make(map[string][]EntityMatcher)
	for _, url := range peerUrls {
		args := strings.Split(url, ":")
		if len(args) != 3 {
			return errors.Errorf("the peer url should be url:port:ip, but got %s", url)
		}
		_, orgName, domain := util.SplitNameOrgDomain(args[0])
		// peer
		peers[args[0]] = Peer{
			URL: fmt.Sprintf("%s:%s", args[0], args[1]),
			GRPCOptions: GRPCOptions{
				SSLTargetNameOverride: args[0],
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
				KeepAlivePermit:       false,
				FailFast:              false,
				AllowInsecure:         false,
			},
			TLSCACerts: TLSCACert{Path: filepath.Join(filePath, defaultCryptoConfigDirName, "peerOrganizations", domain,
				"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", domain))},
		}
		// channel
		channel.Peers[args[0]] = ChannelPeer{
			EndorsingPeer:  true,
			ChaincodeQuery: true,
			LedgerQuery:    true,
			EventSource:    true,
		}

		// organization
		org, exist := organizations[orgName]
		if exist {
			org.Peers = append(org.Peers, args[0])
		} else {
			organizations[orgName] = Organization{
				MSPId:      orgName,
				CryptoPath: fmt.Sprintf("peerOrganizations/%[1]s/users/{username}@%[1]s/msp", domain),
				Peers:      []string{args[0]},
			}
		}

		// entity matcher
		em := EntityMatcher{
			Pattern:                             args[0],
			UrlSubstitutionExp:                  fmt.Sprintf("%s:%s", args[2], args[1]),
			SSLTargetOverrideUrlSubstitutionExp: args[0],
			MappedHost:                          args[0],
		}
		if args[2] == "localhost" || args[2] == "127.0.0.1" {
			em.UrlSubstitutionExp = fmt.Sprintf("%s:%s", args[0], args[1])
		}
		entityMatchers["peer"] = append(entityMatchers["peer"], em)
	}
	connProfile.Peers = peers
	connProfile.Channels = map[string]Channel{
		channelId: channel,
	}

	orderers := make(map[string]Orderer)
	for _, url := range ordererUrls {
		args := strings.Split(url, ":")
		if len(args) != 3 {
			return errors.Errorf("the orderer url should be url:port:ip, but got %s", url)
		}
		_, orgName, domain := util.SplitNameOrgDomain(args[0])
		orderers[args[0]] = Orderer{
			URL: fmt.Sprintf("%s:%s", args[0], args[1]),
			GRPCOptions: GRPCOptions{
				SSLTargetNameOverride: args[0],
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
				KeepAlivePermit:       false,
				FailFast:              false,
				AllowInsecure:         false,
			},
			TLSCACerts: TLSCACert{Path: filepath.Join(filePath, defaultCryptoConfigDirName, "ordererOrganizations", domain,
				"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", domain))},
		}

		_, exist := organizations[orgName]
		if !exist {
			organizations[orgName] = Organization{
				MSPId:      orgName,
				CryptoPath: fmt.Sprintf("ordererOrganizations/%[1]s/users/{username}@%[1]s/msp", domain),
			}
		}

		em := EntityMatcher{
			Pattern:                             args[0],
			UrlSubstitutionExp:                  fmt.Sprintf("%s:%s", args[2], args[1]),
			SSLTargetOverrideUrlSubstitutionExp: args[0],
			MappedHost:                          args[0],
		}
		if args[2] == "localhost" || args[2] == "127.0.0.1" {
			em.UrlSubstitutionExp = fmt.Sprintf("%s:%s", args[0], args[1])
		}
		entityMatchers["orderer"] = append(entityMatchers["orderer"], em)
	}
	connProfile.Orderers = orderers
	connProfile.Organizations = organizations
	connProfile.EntityMatchers = entityMatchers

	data, err := yaml.Marshal(connProfile)
	if err != nil {
		return errors.Wrap(err, "failed to marshal connProfile")
	}
	filePath = filepath.Join(filePath, defaultConnProfileName)
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write connection profile")
	}
	logger.Info("finish generating fabric-sdk-go connection profile")
	return nil
}
