#!/bin/sh

./fdt network generate bootstrap -d /opt/fdt -p peer.testpeerorg1:7051@@127.0.0.1:22: -p peer.testpeerorg2:8051@@127.0.0.1:22: -o orderer.testordererorg:7050@@127.0.0.1:22: -c mychannel -C solo --couchdb
sleep 5
./fdt network startup -d /opt/fdt --couchdb --startonly -c mychannel

cd /opt/fdt && docker-compose -f docker-compose-orderer-testordererorg.yaml -f docker-compose-peer-testpeerorg1-couchdb.yaml -f docker-compose-peer-testpeerorg1.yaml -f docker-compose-peer-testpeerorg2.yaml -f docker-compose-peer-testpeerorg2-couchdb.yaml down -v