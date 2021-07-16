package sdkutil

import (
	"fmt"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"path/filepath"
)

var logger = mylogger.NewLogger()

// 与fabric-sdk-go相关的方法

// TODO 安装链码，实例化链码，更新链码

const defaultUsername = "Admin"

type FabricSDKDriver struct {
	connProfilePath string
	fabSDK         *fabsdk.FabricSDK
}

// NewFabricSDKDriver creates a fabric-sdk-go instance using specified connection profile
func NewFabricSDKDriver(connProfilePath string) (*FabricSDKDriver, error) {
	sdk, err := fabsdk.New(config.FromFile(connProfilePath))
	if err != nil {
		return nil, err
	}
	return &FabricSDKDriver{connProfilePath, sdk}, nil
}

// CreateChannel 使用sdk创建指定名称的通道
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
		ChannelID: channelId,
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

func (driver *FabricSDKDriver) JoinChannel(channelId string, orgId string, ordererEndpoint string) error {
	adminContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		return errors.Wrap(err, "failed to create new resource management client")
	}
	if err = orgResMgmt.JoinChannel(channelId, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(ordererEndpoint)); err != nil {
		return errors.Wrapf(err, "%s peers failed to join channel %s", orgId, channelId)
	}
	logger.Infof("org %s peers join channel %s success", orgId, channelId)
	return nil
}

func (driver *FabricSDKDriver) InstallChaincode() error {
	return nil
}

func (driver *FabricSDKDriver) InstantiateChaincode() error {
	return nil
}

func (driver *FabricSDKDriver) UpdateChaincode() error {
	return nil
}

func (driver *FabricSDKDriver) Close() {
	driver.fabSDK.Close()
}