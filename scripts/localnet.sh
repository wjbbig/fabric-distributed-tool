#!/bin/sh

./fdt network generate --bootstrap -d /opt/fdt -p peer.testpeerorg1:7051@@127.0.0.1:22: -p peer.testpeerorg2:8051@@127.0.0.1:22: -o orderer.testordererorg:7050@@127.0.0.1:22: -c mychannel -C solo \
-n mycc -P "/home/wjbbig/workspace/go/src/github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go" -v v1 -r "OR('testpeerorg1.peer','testpeerorg2.peer')" -f "InitLedger" -w testnetwork -V v2.0 --initcc
echo "sleep 5s"
sleep 5
./fdt network startup -d /opt/fdt
echo "sleep 5s"
sleep 5

./fdt network generate -d /opt/fdt -p peer1.testpeerorg1:9051@@127.0.0.1:22: -p peer.testpeerorg3:10051@@127.0.0.1:22:--extend

./fdt network startnode -d /opt/fdt -n peer1.testpeerorg1

./fdt network existorgjoin -d /opt/fdt -c mychannel -n peer1.testpeerorg1

./fdt network neworgjoin -d /opt/fdt -c mychannel -n peer.testpeerorg3

# deploycc cmd
./fdt network deploycc -d /opt/fdt -c mychannel -n nsf -p "/home/wjbbig/workspace/go/src/github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go" -v v1 -f InitLedger -P "OR('testpeerorg1.peer','testpeerorg2.peer')" --initcc

./fdt network upgradecc -d /opt/fdt -c mychannel -n mycc -p "/home/wjbbig/workspace/go/src/github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go" -v v2 -f InitLedger -P "OR('testpeerorg1.peer')" --initcc --redeploy

./fdt network stopnode -d /opt/fdt -n peer.testpeerorg1

echo "shutdown the fabric network"
./fdt network shutdown -d /opt/fdt