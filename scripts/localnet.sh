#!/bin/sh

./fdt network generate bootstrap -d /opt/fdt -p peer.testpeerorg1:7051@@127.0.0.1:22: -p peer.testpeerorg2:8051@@127.0.0.1:22: -o orderer.testordererorg:7050@@127.0.0.1:22: -c mychannel -C solo --couchdb