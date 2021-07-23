package startup

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"github.com/wjbbig/fabric-distributed-tool/util"
	"path/filepath"
	"strings"
)

var logger = mylogger.NewLogger()
var (
	dataDir       string
	ca            bool
	stateDB       string
	channelId     string
	chaincodeId   string
	chaincodePath string
	initParam     string
	startOnly     bool
)

func init() {
	resetFlags()
}

// 启动fabric网络
var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "start the fabric network",
	Long:  "start the fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			logger.Error("datadir is not specified")
		}
		sshUtil, err := readSSHConfig()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		defer sshUtil.CloseAll()

		if err := transferFilesByPeerName(sshUtil); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := startupNetwork(sshUtil); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// if only starting the fabric docker container
		if startOnly {
			return nil
		}

		// todo
		return nil
	},
}

func Cmd() *cobra.Command {
	startupCmd.Flags().AddFlagSet(flags)
	return startupCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVar(&dataDir, "datadir", "", "Path to file containing fabric needed")
	flags.BoolVar(&ca, "ca", false, "If start the fabric ca container")
	flags.StringVar(&stateDB, "statedb", "leveldb", "Which type of statedb that fabric uses")
	flags.BoolVar(&startOnly, "startonly", false, "Just start the docker container")
	flags.StringVar(&channelId, "channel", "", "Channel Name")
	flags.StringVar(&chaincodeId, "chaincodeid", "", "Chaincode Name")
	flags.StringVar(&chaincodePath, "path", "", "Chaincode path")
	flags.StringVar(&initParam, "initparam", "", "Chaincode Init params")
}

func readSSHConfig() (*sshutil.SSHUtil, error) {
	sshConfig, err := sshutil.UnmarshalSSHConfig(dataDir)
	if err != nil {
		return nil, err
	}
	sshUtil := sshutil.NewSSHUtil()
	for _, client := range sshConfig.Clients {
		if err := sshUtil.Add(client.Name, client.Username, client.Password, fmt.Sprintf("%s:%s", client.Host, client.Port), client.Type); err != nil {
			return nil, err
		}
	}
	return sshUtil, nil
}

func transferFilesByPeerName(sshUtil *sshutil.SSHUtil) error {
	ordererCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "ordererOrganizations")
	peerCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "peerOrganizations")
	for name, client := range sshUtil.Clients() {
		_, orgName, _ := util.SplitNameOrgDomain(name)
		// send node self keypairs and certs
		var certDir string
		if client.GetNodeType() == "peer" {
			certDir = filepath.Join(peerCryptoConfigPrefix, orgName)
		} else {
			certDir = filepath.Join(ordererCryptoConfigPrefix, orgName)
		}
		err := client.Sftp(certDir, certDir)
		if err != nil {
			return err
		}
		// send genesis.block, channel.tx and anchor.tx
		channelArtifactsPath := filepath.Join(dataDir, "channel-artifacts")
		if err = client.Sftp(channelArtifactsPath, channelArtifactsPath); err != nil {
			return err
		}

		dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
		if err = client.Sftp(dockerComposeFilePath, dataDir); err != nil {
			return err
		}

		// todo transfer chaincode files
	}
	return nil
}

func startupNetwork(sshUtil *sshutil.SSHUtil) error {
	for name, client := range sshUtil.Clients() {
		dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
		// start node
		if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
			return err
		}
		// TODO start ca if chosen

		// TODO start couchdb if chosen
	}
	return nil
}
