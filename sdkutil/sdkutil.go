package sdkutil

import (
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"log"
)

//var log

// 与fabric-sdk-go相关的方法

// TODO 安装链码，实例化链码，更新链码

const defaultUsername = "Admin"

type FabricSDKDriver struct {
	connProfilePath string
	fabSDK          map[string]*fabsdk.FabricSDK
}

// NewFabricSDKDriver
func NewFabricSDKDriver(connProfilePath string, channelId string) (*FabricSDKDriver, error) {
	sdk, err := fabsdk.New(config.FromFile(connProfilePath))
	if err != nil {
		return nil, err
	}
	sdks := make(map[string]*fabsdk.FabricSDK)
	sdks[channelId] = sdk
	return &FabricSDKDriver{connProfilePath, sdks}, nil
}

// CreateChannel 使用sdk创建指定名称的通道
func (driver *FabricSDKDriver) CreateChannel(channelId string, orgId string) error {
	sdk, exist := driver.fabSDK[channelId]
	if !exist {
		return errors.Errorf("connProfile file not exists. name=%s", channelId)
	}
	clientContext := sdk.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}

	return createChannel(sdk, resMgmtClient, channelId, orgId)
}

func createChannel(sdk *fabsdk.FabricSDK, resMgmtClient *resmgmt.Client, channelId, orgId string) error {
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
		// TODO 补充这里的路径
		ChannelConfigPath: "",
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}
	resp, err := resMgmtClient.SaveChannel(createChannelReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		// TODO 这里需要输入一个orderer的名字
		resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		return err
	}
	log.Printf("create channel %s success, txId=%s", channelId, resp.TransactionID)
	return nil
}

func (driver *FabricSDKDriver) JoinChannel(channelId string, orgId string) error {
	sdk, exist := driver.fabSDK[channelId]
	if !exist {
		return errors.Errorf("connProfile file not exists. name=%s", channelId)
	}
	adminContext := sdk.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		return errors.Wrap(err, "failed to create new resource management client")
	}
	if err = orgResMgmt.JoinChannel(channelId, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint("")); err != nil {
		return errors.Wrapf(err, "%s peers failed to join channel %s", orgId, channelId)
	}
	log.Printf("org %s peers join channel %s success", orgId, channelId)
	return nil
}
