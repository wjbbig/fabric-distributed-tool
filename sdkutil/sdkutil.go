package sdkutil

import (
	"fmt"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"os"
	"path/filepath"
)

var logger = mylogger.NewLogger()

// 与fabric-sdk-go相关的方法

// TODO 安装链码，实例化链码，更新链码

const defaultUsername = "Admin"

type FabricSDKDriver struct {
	connProfilePath string
	fabSDK          *fabsdk.FabricSDK
}

// NewFabricSDKDriver creates a fabric-sdk-go instance using specified connection profile
func NewFabricSDKDriver(connProfilePath string) (*FabricSDKDriver, error) {
	sdk, err := fabsdk.New(config.FromFile(connProfilePath))
	if err != nil {
		return nil, err
	}
	return &FabricSDKDriver{connProfilePath, sdk}, nil
}

// CreateChannel creates a channel with specified channelId
func (driver *FabricSDKDriver) CreateChannel(channelId string, orgId string, fileDir string, ordererEndpoint string) error {
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}

	return createChannel(driver.fabSDK, resMgmtClient, channelId, orgId, fileDir, ordererEndpoint)
}

func createChannel(sdk *fabsdk.FabricSDK, resMgmtClient *resmgmt.Client, channelId, orgId, fileDir, ordererEndpoint string) error {
	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orgId))
	if err != nil {
		return err
	}
	adminIdentity, err := mspClient.GetSigningIdentity(defaultUsername)
	if err != nil {
		return err
	}
	createChannelReq := resmgmt.SaveChannelRequest{
		ChannelID:         channelId,
		ChannelConfigPath: filepath.Join(fileDir, "channel-artifacts", fmt.Sprintf("%s.tx", channelId)),
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}
	resp, err := resMgmtClient.SaveChannel(createChannelReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(ordererEndpoint))
	if err != nil {
		return err
	}
	logger.Infof("create channel %s success, txId=%s", channelId, resp.TransactionID)
	return nil
}

func (driver *FabricSDKDriver) JoinChannel(channelId string, orgId string, ordererEndpoint string, peerEndpoint string) error {
	adminContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		return errors.Wrap(err, "failed to create new resource management client")
	}
	if err = orgResMgmt.JoinChannel(channelId,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(ordererEndpoint),
		resmgmt.WithTargetEndpoints(peerEndpoint)); err != nil {
		return errors.Wrapf(err, "%s peers failed to join channel %s", orgId, channelId)
	}
	logger.Infof("peer %s joins channel %s success", orgId, channelId)
	return nil
}

func (driver *FabricSDKDriver) InstallCC(ccId, ccPath, ccVersion, channelId, orgId, peerEndpoint string) error {
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}
	gopath := os.Getenv("GOPATH")
	ccPkg, err := gopackager.NewCCPackage(ccPath, gopath)
	if err != nil {
		return errors.Wrapf(err, "package chaincode failed")
	}
	installCCReq := resmgmt.InstallCCRequest{Name: ccId, Path: ccPath, Version: ccVersion, Package: ccPkg}
	_, err = resMgmtClient.InstallCC(installCCReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(peerEndpoint))
	if err != nil {
		return errors.Wrapf(err, "install chaincode failed")
	}
	logger.Infof("install chaincode %s on %s success", ccId, orgId)
	return nil
}

func (driver *FabricSDKDriver) InstantiateCC(ccId, ccPath, ccVersion, channelId, orgId, policy, peerEndpoint string, initArgs []string) error {
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}
	ccPolicy, err := policydsl.FromString(policy)
	if err != nil {
		return errors.Wrap(err, "unmarshal policy string failed")
	}
	response, err := resMgmtClient.InstantiateCC(
		channelId,
		resmgmt.InstantiateCCRequest{Name: ccId, Path: ccPath, Version: ccVersion, Args: parseCCArgs(initArgs), Policy: ccPolicy},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(peerEndpoint),
	)
	if err != nil {
		return errors.Wrap(err, "instantiate chaincode failed")
	}
	logger.Infof("instantiate chaincode %s success, txid=%s", ccId, response.TransactionID)
	return nil
}

func (driver *FabricSDKDriver) UpdateCC(ccId, ccPath, ccVersion, channelId, orgId, policy string, initArgs []string) error {
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}
	ccPolicy, err := policydsl.FromString(policy)
	if err != nil {
		return errors.Wrap(err, "unmarshal policy string failed")
	}
	req := resmgmt.UpgradeCCRequest{
		Name:    ccId,
		Path:    ccPath,
		Version: ccVersion,
		Args:    parseCCArgs(initArgs),
		Policy:  ccPolicy,
	}
	resp, err := resMgmtClient.UpgradeCC(channelId, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrapf(err, "upgrade chaincode %s failed, err=%s", ccId, err)
	}
	logger.Infof("upgrade chaincode %s success, txid=%s", ccId, resp.TransactionID)
	return nil
}

func (driver *FabricSDKDriver) Close() {
	driver.fabSDK.Close()
}

func parseCCArgs(args []string) [][]byte {
	var argBytes [][]byte
	for _, arg := range args {
		argBytes = append(argBytes, []byte(arg))
	}
	return argBytes
}
