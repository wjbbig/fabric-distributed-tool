package sdkutil

import (
	"fmt"
	"path/filepath"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/pkg/errors"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/utils"
)

var logger = mylogger.NewLogger()

// utils of fabric-sdk-go

type SDKUtil interface {
	CreateChannel(channelId string, orgId string, fileDir string, ordererEndpoint string, peerEndpoint string) error
	JoinChannel(channelId string, orgId string, ordererEndpoint string, peerEndpoint string) error
	InstallCC(ccId, ccPath, ccVersion, channelId, orgId, peerEndpoint string) error
	InstantiateCC(ccId, ccPath, ccVersion, channelId, orgId, policy, peerEndpoint string, initArgs []string) error
	UpdateCC(ccId, ccPath, ccVersion, channelId, orgId, policy string, initArgs []string) error
	Close()
}

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
func (driver *FabricSDKDriver) CreateChannel(channelId string, orgId string, fileDir string, ordererEndpoint string, peerEndpoint string) error {
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}

	return createChannel(driver.fabSDK, resMgmtClient, channelId, orgId, fileDir, ordererEndpoint, peerEndpoint)
}

func createChannel(sdk *fabsdk.FabricSDK, resMgmtClient *resmgmt.Client, channelId, orgId, fileDir, ordererEndpoint, peerEndpoint string) error {
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
		resmgmt.WithOrdererEndpoint(ordererEndpoint),
		resmgmt.WithTargetEndpoints(peerEndpoint))
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
	ccPkg, err := gopackager.NewCCPackage(ccPath, "")
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

//================================v2.0====================================

func (driver *FabricSDKDriver) PackageCC(ccId, ccPath, outputPath string) (string, []byte, error) {
	desc := &lcpackager.Descriptor{
		Type:  pb.ChaincodeSpec_GOLANG,
		Path:  ccPath,
		Label: ccId,
	}
	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		return "", nil, errors.Wrap(err, "error packaging chaincode")
	}
	if err := utils.WriteFile(outputPath, ccPkg, 0755); err != nil {
		return "", nil, errors.Wrap(err, "error writing chaincode")
	}
	return desc.Label, ccPkg, nil
}

// InstallCCV2 uses to install chaincode for fabric v2.0
func (driver *FabricSDKDriver) InstallCCV2(ccId, channelId, orgId, peerEndpoint string, ccPkg []byte) error {
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   ccId,
		Package: ccPkg,
	}
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}
	packageId := lcpackager.ComputePackageID(ccId, ccPkg)
	resp, err := resMgmtClient.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrap(err, "install chaincode failed")
	}
	if resp[0].PackageID != packageId {
		return errors.Wrap(err, "packageId from install response is not equal to the id from computed")
	}

	return nil
}

func (driver *FabricSDKDriver) QueryInstalled(ccId, channelId, orgId, peerEndpoint, packageId string) error {
	return nil
}

func (driver *FabricSDKDriver) ApproveCC(ccId, ccVersion, ccPolicy, channelId, orgId, peerEndpoint, ordererEndpoint, packageId string) error {
	policy, err := policydsl.FromString(ccPolicy)
	if err != nil {
		return errors.Wrap(err, "build ccPolicy failed")
	}
	clientContext := driver.fabSDK.Context(fabsdk.WithUser(defaultUsername), fabsdk.WithOrg(orgId))
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrapf(err, "create resmgmt client failed, channel name=%s", channelId)
	}
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:      ccId,
		Version:   ccVersion,
		PackageID: packageId,
		// TODO: sequence
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   policy,
	}

	txId, err := resMgmtClient.LifecycleApproveCC(channelId, approveCCReq, resmgmt.WithTargetEndpoints(peerEndpoint),
		resmgmt.WithOrdererEndpoint(ordererEndpoint), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrap(err, "approve chaincode failed")
	}
	logger.Infof("approve chaincode %s success for org %s, txId=%s", ccId, orgId, txId)
	return nil
}

func (driver *FabricSDKDriver) QueryApproveCC(ccId string) error {
	return nil
}

func (driver *FabricSDKDriver) CommitCC() error {
	// ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	// req := resmgmt.LifecycleCommitCCRequest{
	// 	Name:              ccID,
	// 	Version:           "0",
	// 	Sequence:          1,
	// 	EndorsementPlugin: "escc",
	// 	ValidationPlugin:  "vscc",
	// 	SignaturePolicy:   ccPolicy,
	// 	InitRequired:      true,
	// }
	// txnID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithTargetEndpoints(peer1), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// require.NotEmpty(t, txnID)
	return nil
}

func (driver *FabricSDKDriver) QueryCommittedCC() error {

	return nil
}

func (driver *FabricSDKDriver) InitCC() error {
	return nil
}
