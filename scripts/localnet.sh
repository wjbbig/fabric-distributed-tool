#!/bin/sh

./fdt network generate bootstrap -d /opt/fdt -p peer.testpeerorg1:7051@@127.0.0.1:22: -p peer.testpeerorg2:8051@@127.0.0.1:22: -o orderer.testordererorg:7050@@127.0.0.1:22: -c mychannel -C solo --couchdb \
-n mycc -P github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go -v v1 -r "OR('testpeerorg1.peer','testpeerorg2.peer')" -i "init,a,100,b,100" -w testnetwork
sleep 5
./fdt network startup -d /opt/fdt --startonly

# cd /opt/fdt && docker-compose -f docker-compose-orderer-testordererorg.yaml -f docker-compose-peer-testpeerorg1-couchdb.yaml -f docker-compose-peer-testpeerorg1.yaml -f docker-compose-peer-testpeerorg2.yaml -f docker-compose-peer-testpeerorg2-couchdb.yaml down -v