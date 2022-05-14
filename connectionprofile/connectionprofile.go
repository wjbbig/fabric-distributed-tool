package connectionprofile

import (
	"fmt"
	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"github.com/wjbbig/fabric-distributed-tool/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	DefaultConnProfileName     = "connection-config.yaml"
	defaultCryptoConfigDirName = "crypto-config"
)

var logger = mylogger.NewLogger()

// ConnProfile generates connection profile for fabric-sdk-go
type ConnProfile struct {
	Version                string                           `yaml:"version,omitempty"`
	Client                 *Client                          `yaml:"client,omitempty"`
	Channels               map[string]*Channel              `yaml:"channels,omitempty"`
	Organizations          map[string]*Organization         `yaml:"organizations,omitempty"`
	Orderers               map[string]*Orderer              `yaml:"orderers,omitempty"`
	Peers                  map[string]*Peer                 `yaml:"peers,omitempty"`
	CertificateAuthorities map[string]*CertificateAuthority `yaml:"certificateAuthorities,omitempty"`
	EntityMatchers         map[string][]*EntityMatcher      `yaml:"entityMatchers,omitempty"`
}

// ============================CLIENT==================================

type Client struct {
	Organization    string           `yaml:"organization,omitempty"`
	Logger          *Logger          `yaml:"logger,omitempty"`
	CryptoConfig    *CryptoConfig    `yaml:"cryptoconfig,omitempty"`
	CredentialStore *CredentialStore `yaml:"credentialStore,omitempty"`
	BCCSP           *BCCSP           `yaml:"BCCSP,omitempty"`
	TLSCerts        *TLSCerts        `yaml:"tlsCerts,omitempty"`
}

type Logger struct {
	Level string `yaml:"level,omitempty"`
}

type CryptoConfig struct {
	Path string `yaml:"path,omitempty"`
}

type CredentialStore struct {
	Path        string       `yaml:"path,omitempty"`
	CryptoStore *CryptoStore `yaml:"cryptoStore,omitempty"`
}

type CryptoStore struct {
	Path string `yaml:"path,omitempty"`
}

type BCCSP struct {
	Security *BCCSPSecurity `yaml:"security,omitempty"`
}

type BCCSPSecurity struct {
	Enable        bool          `yaml:"enable,omitempty"`
	Default       *BCCSPDefault `yaml:"default,omitempty"`
	HashAlgorithm string        `yaml:"hashAlgorithm,omitempty"`
	SoftVerify    bool          `yaml:"softVerify,omitempty"`
	Level         int           `yaml:"level,omitempty"`
}

type BCCSPDefault struct {
	Provider string `yaml:"provider,omitempty"`
}

type TLSCerts struct {
	SystemCertPool bool       `yaml:"systemCertPool,omitempty"`
	TLSClient      *TLSClient `yaml:"client,omitempty"`
}

type TLSClient struct {
	Key  *Key  `yaml:"key,omitempty"`
	Cert *Cert `yaml:"cert,omitempty"`
}

type Key struct {
	Path string `yaml:"path,omitempty"`
}

type Cert struct {
	Path string `yaml:"path,omitempty"`
}

// ========================CHANNEL===========================

type Channel struct {
	Peers    map[string]*ChannelPeer `yaml:"peers,omitempty"`
	Policies *ChannelPolicies        `yaml:"policies,omitempty"`
}

type ChannelPeer struct {
	EndorsingPeer  bool `yaml:"endorsingPeer,omitempty"`
	ChaincodeQuery bool `yaml:"chaincodeQuery,omitempty"`
	LedgerQuery    bool `yaml:"ledgerQuery,omitempty"`
	EventSource    bool `yaml:"eventSource,omitempty"`
}

type ChannelPolicies struct {
	QueryChannelConfig *QueryChannelConfig `yaml:"queryChannelConfig,omitempty"`
	Discovery          *Discovery          `yaml:"discovery,omitempty"`
	EventService       *EventService       `yaml:"eventService,omitempty"`
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
	URL         string       `yaml:"url,omitempty"`
	GRPCOptions *GRPCOptions `yaml:"grpcOptions,omitempty"`
	TLSCACerts  *TLSCACert   `yaml:"tlsCACerts,omitempty"`
}

// ==============================PEER=====================================

type Peer struct {
	URL         string       `yaml:"url,omitempty"`
	GRPCOptions *GRPCOptions `yaml:"grpcOptions,omitempty"`
	TLSCACerts  *TLSCACert   `yaml:"tlsCACerts,omitempty"`
}

type GRPCOptions struct {
	SSLTargetNameOverride string `yaml:"ssl-target-name-override,omitempty"`
	KeepAliveTime         string `yaml:"keep-alive-time,omitempty"`
	KeepAliveTimeout      string `yaml:"keep-alive-timeout,omitempty"`
	KeepAlivePermit       bool   `yaml:"keep-alive-permit"`
	FailFast              bool   `yaml:"fail-fast"`
	AllowInsecure         bool   `yaml:"allow-insecure"`
}

// ============================CA=================================

type TLSCACert struct {
	Path      string     `yaml:"path,omitempty"`
	TLSClient *TLSClient `yaml:"client,omitempty"`
}

type Registrar struct {
	EnrollId     string `yaml:"enrollId,omitempty"`
	EnrollSecret string `yaml:"enrollSecret,omitempty"`
}

type CertificateAuthority struct {
	URL       string     `yaml:"url,omitempty"`
	Registrar *Registrar `yaml:"registrar,omitempty"`
	TLSCACert *TLSCACert `yaml:"tlsCACerts,omitempty"`
	CAName    string     `yaml:"caName,omitempty"`
}

// =============================ENTITYMATCHER================================

type EntityMatcher struct {
	Pattern                             string `yaml:"pattern,omitempty"`
	UrlSubstitutionExp                  string `yaml:"urlSubstitutionExp,omitempty"`
	SSLTargetOverrideUrlSubstitutionExp string `yaml:"sslTargetOverrideUrlSubstitutionExp,omitempty"`
	MappedHost                          string `yaml:"mappedHost,omitempty"`
}

type ChannelEntity struct {
	ChannelId string
	Peers     []string
	Orderers  []string
}

// GenerateNetworkConnProfile generates the connection profile, the format peer or orderer must be 'url:port:ip'
func GenerateNetworkConnProfile(filePath string, channelEntities []ChannelEntity) error {
	logger.Info("begin to generate fabric-sdk-go connection profile")
	var connProfile ConnProfile
	connProfile.Version = "1.0.0"
	client := &Client{
		Logger:       &Logger{"error"},
		CryptoConfig: &CryptoConfig{filepath.Join(filePath, defaultCryptoConfigDirName)},
		CredentialStore: &CredentialStore{
			Path:        filepath.Join(filePath, "state-store"),
			CryptoStore: &CryptoStore{filepath.Join(filePath, "msp")},
		},
		BCCSP: &BCCSP{&BCCSPSecurity{
			Enable:        true,
			Default:       &BCCSPDefault{"SW"},
			HashAlgorithm: "SHA2",
			SoftVerify:    true,
			Level:         256,
		}},
		TLSCerts: &TLSCerts{
			SystemCertPool: false,
		},
	}
	connProfile.Client = client

	peers := make(map[string]*Peer)
	organizations := make(map[string]*Organization)
	entityMatchers := make(map[string][]*EntityMatcher)
	connProfile.Channels = map[string]*Channel{}
	var peerOrgName string
	for _, entity := range channelEntities {
		channel := &Channel{Peers: map[string]*ChannelPeer{}}
		for _, url := range entity.Peers {
			args := strings.Split(url, ":")
			if len(args) != 3 {
				return errors.Errorf("the peer url should be url:port:ip, but got %s", url)
			}
			_, orgName, domain := utils.SplitNameOrgDomain(args[0])
			peerOrgName = orgName
			// channel
			channel.Peers[args[0]] = &ChannelPeer{
				EndorsingPeer:  true,
				ChaincodeQuery: true,
				LedgerQuery:    true,
				EventSource:    true,
			}
			// peer
			if _, ok := peers[args[0]]; !ok {
				peers[args[0]] = &Peer{
					URL: fmt.Sprintf("%s:%s", args[0], args[1]),
					GRPCOptions: &GRPCOptions{
						SSLTargetNameOverride: args[0],
						KeepAliveTime:         "0s",
						KeepAliveTimeout:      "20s",
						KeepAlivePermit:       false,
						FailFast:              false,
						AllowInsecure:         false,
					},
					TLSCACerts: &TLSCACert{Path: filepath.Join(filePath, defaultCryptoConfigDirName, "peerOrganizations", domain,
						"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", domain))},
				}

				// organization
				org, exist := organizations[orgName]
				if exist {
					org.Peers = append(org.Peers, args[0])
				} else {
					organizations[orgName] = &Organization{
						MSPId:      orgName,
						CryptoPath: fmt.Sprintf("peerOrganizations/%[1]s/users/{username}@%[1]s/msp", domain),
						Peers:      []string{args[0]},
					}
				}

				// entity matcher
				em := &EntityMatcher{
					Pattern:                             args[0],
					UrlSubstitutionExp:                  fmt.Sprintf("%s:%s", args[2], args[1]),
					SSLTargetOverrideUrlSubstitutionExp: args[0],
					MappedHost:                          args[0],
				}
				entityMatchers["peer"] = append(entityMatchers["peer"], em)
			}
		}
		connProfile.Channels[entity.ChannelId] = channel
	}
	client.Organization = peerOrgName
	connProfile.Peers = peers

	orderers := make(map[string]*Orderer)
	for _, entity := range channelEntities {
		for _, url := range entity.Orderers {
			args := strings.Split(url, ":")
			if len(args) != 3 {
				return errors.Errorf("the orderer url should be url:port:ip, but got %s", url)
			}
			_, orgName, domain := utils.SplitNameOrgDomain(args[0])
			if _, ok := orderers[args[0]]; !ok {
				orderers[args[0]] = &Orderer{
					URL: fmt.Sprintf("%s:%s", args[0], args[1]),
					GRPCOptions: &GRPCOptions{
						SSLTargetNameOverride: args[0],
						KeepAliveTime:         "0s",
						KeepAliveTimeout:      "20s",
						KeepAlivePermit:       false,
						FailFast:              false,
						AllowInsecure:         false,
					},
					TLSCACerts: &TLSCACert{Path: filepath.Join(filePath, defaultCryptoConfigDirName, "ordererOrganizations", domain,
						"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", domain))},
				}

				_, exist := organizations[orgName]
				if !exist {
					organizations[orgName] = &Organization{
						MSPId:      orgName,
						CryptoPath: fmt.Sprintf("ordererOrganizations/%[1]s/users/{username}@%[1]s/msp", domain),
					}
				}

				em := &EntityMatcher{
					Pattern:                             args[0],
					UrlSubstitutionExp:                  fmt.Sprintf("%s:%s", args[2], args[1]),
					SSLTargetOverrideUrlSubstitutionExp: args[0],
					MappedHost:                          args[0],
				}
				entityMatchers["orderer"] = append(entityMatchers["orderer"], em)
			}
		}
	}
	connProfile.Orderers = orderers
	connProfile.Organizations = organizations
	connProfile.EntityMatchers = entityMatchers

	data, err := yaml.Marshal(connProfile)
	if err != nil {
		return errors.Wrap(err, "failed to marshal connProfile")
	}
	filePath = filepath.Join(filePath, DefaultConnProfileName)
	if err := utils.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write connection profile")
	}
	logger.Info("finish generating fabric-sdk-go connection profile")
	return nil
}

func UnmarshalConnectionProfile(dataDir string) (*ConnProfile, error) {
	filePath := filepath.Join(dataDir, DefaultConnProfileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "connection profile does not exist")
	}
	profile := &ConnProfile{}
	if err = yaml.Unmarshal(data, profile); err != nil {
		return nil, errors.Wrapf(err, "yaml unmarshal failed")
	}

	return profile, nil
}

func (profile *ConnProfile) ExtendChannel(dataDir, channelId string, peers []string) error {
	_, exist := profile.Channels[channelId]
	if exist {
		return errors.Errorf("channel %s exists", channelId)
	}

	channel := &Channel{Peers: make(map[string]*ChannelPeer)}
	for _, peer := range peers {
		channel.Peers[peer] = &ChannelPeer{
			EndorsingPeer:  true,
			ChaincodeQuery: true,
			LedgerQuery:    true,
			EventSource:    true,
		}
	}

	profile.Channels[channelId] = channel

	data, err := yaml.Marshal(profile)
	if err != nil {
		return errors.Wrap(err, "failed to marshal connProfile")
	}
	filePath := filepath.Join(dataDir, DefaultConnProfileName)
	if err := utils.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "failed to write connection profile")
	}
	return nil
}

func (profile *ConnProfile) ExtendNodesAndOrgs(dataDir string, peers, orderers []*network.Node) error {

	for _, node := range peers {
		peer := &Peer{
			URL: fmt.Sprintf("%s:%d", node.GetHostname(), node.NodePort),
			GRPCOptions: &GRPCOptions{
				SSLTargetNameOverride: node.GetHostname(),
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
			},
			TLSCACerts: &TLSCACert{Path: filepath.Join(dataDir, defaultCryptoConfigDirName, "peerOrganizations", node.Domain,
				"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", node.Domain))},
		}
		profile.Peers[node.GetHostname()] = peer

		peerEntityMatcher := &EntityMatcher{
			Pattern:                             node.GetHostname(),
			UrlSubstitutionExp:                  fmt.Sprintf("%s:%d", node.Host, node.NodePort),
			SSLTargetOverrideUrlSubstitutionExp: node.GetHostname(),
			MappedHost:                          node.GetHostname(),
		}

		profile.EntityMatchers["peer"] = append(profile.EntityMatchers["peer"], peerEntityMatcher)

		org, exist := profile.Organizations[node.OrgId]
		if !exist {
			org := &Organization{
				MSPId:      node.OrgId,
				CryptoPath: fmt.Sprintf("peerOrganizations/%[1]s/users/{username}@%[1]s/msp", node.Domain),
				Peers:      []string{node.GetHostname()},
			}
			profile.Organizations[node.OrgId] = org
		} else {
			org.Peers = append(org.Peers, node.GetHostname())
		}
	}
	for _, node := range orderers {
		orderer := &Orderer{
			URL: fmt.Sprintf("%s:%d", node.GetHostname(), node.NodePort),
			GRPCOptions: &GRPCOptions{
				SSLTargetNameOverride: node.GetHostname(),
				KeepAliveTime:         "0s",
				KeepAliveTimeout:      "20s",
			},
			TLSCACerts: &TLSCACert{Path: filepath.Join(dataDir, defaultCryptoConfigDirName, "ordererOrganizations", node.Domain,
				"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", node.Domain))},
		}
		profile.Orderers[node.GetHostname()] = orderer

		ordererEntityMatcher := &EntityMatcher{
			Pattern:                             node.GetHostname(),
			UrlSubstitutionExp:                  fmt.Sprintf("%s:%d", node.Host, node.NodePort),
			SSLTargetOverrideUrlSubstitutionExp: node.GetHostname(),
			MappedHost:                          node.GetHostname(),
		}

		profile.EntityMatchers["orderer"] = append(profile.EntityMatchers["orderer"], ordererEntityMatcher)

		_, exist := profile.Organizations[node.OrgId]
		if !exist {
			org := &Organization{
				MSPId:      node.OrgId,
				CryptoPath: fmt.Sprintf("ordererOrganizations/%[1]s/users/{username}@%[1]s/msp", node.Domain),
			}
			profile.Organizations[node.OrgId] = org
		}
	}
	data, err := yaml.Marshal(profile)
	if err != nil {
		return errors.Wrap(err, "yaml marshal failed")
	}
	filePath := filepath.Join(dataDir, DefaultConnProfileName)
	if err := utils.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "write file failed")
	}
	return nil
}

func (profile *ConnProfile) ExtendCANodeByOrg(dataDir string, caNode *network.Node, caInfo *network.CA) error {
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	organization, ok := profile.Organizations[caNode.OrgId]
	if !ok {
		return errors.Errorf("org %s does not exist", organization)
	}

	organization.CertificateAuthorities = append(organization.CertificateAuthorities, caNode.Name)

	if profile.CertificateAuthorities == nil {
		profile.CertificateAuthorities = make(map[string]*CertificateAuthority)
	}

	profile.CertificateAuthorities[caNode.Name] = &CertificateAuthority{
		URL: fmt.Sprintf("http://%s:%d", caNode.Name, caNode.NodePort),
		Registrar: &Registrar{
			EnrollId:     caInfo.EnrollId,
			EnrollSecret: caInfo.EnrollSecret,
		},
		TLSCACert: &TLSCACert{
			Path: filepath.Join(dataDir, defaultCryptoConfigDirName, "peerOrganizations", caNode.Domain,
				"tlsca", fmt.Sprintf("tlsca.%s-cert.pem", caNode.Domain)),
			TLSClient: &TLSClient{
				Key: &Key{filepath.Join(dataDir, defaultCryptoConfigDirName, "peerOrganizations", caNode.Domain,
					"users", fmt.Sprintf("User1@%s", caNode.Domain), "tls", "client.key")},
				Cert: &Cert{filepath.Join(dataDir, defaultCryptoConfigDirName, "peerOrganizations", caNode.Domain,
					"users", fmt.Sprintf("User1@%s", caNode.Domain), "tls", "client.crt")},
			},
		},
		CAName: "ca-" + caNode.OrgId,
	}

	profile.EntityMatchers["certificateAuthority"] = append(profile.EntityMatchers["certificateAuthority"], &EntityMatcher{
		Pattern:            caNode.Name,
		UrlSubstitutionExp: fmt.Sprintf("http://%s:%d", caNode.Host, caNode.NodePort),
		MappedHost:         caNode.Name,
	})

	data, err := yaml.Marshal(profile)
	if err != nil {
		return errors.Wrap(err, "yaml marshal failed")
	}
	filePath := filepath.Join(dataDir, DefaultConnProfileName)
	if err := utils.WriteFile(filePath, data, 0755); err != nil {
		return errors.Wrap(err, "write file failed")
	}
	return nil
}
