# fabric-distributed-tool
fabric-distributed-tool is a tool for distributed deployment of hyperledger fabric. It's easy to build up or update a fabric network with this tool. Now it supports fabric v1.4.x and v2.
## Install

---
~~~bash
go build -o fdt cmd/root_cmd.go
~~~
or using Makefile
~~~bash
make build
~~~
## How to use

---
Use the bootstrap command, the form of node is hostname:port@server_username@server_ip:ssh_port:server_password.
~~~bash
fdt network generate --bootstrap -d ./fdtdata -p peer.testpeerorg1:7051@@127.0.0.1:22: -p peer.testpeerorg2:8051@@127.0.0.1:22: -o orderer.testordererorg:7050@@127.0.0.1:22: \
  -c mychannel -C solo -n mycc -P "$GOPATH/github.com/hyperledger/fabric-samples/chaincode/fabcar/go" -v v1 -r "OR('testpeerorg1.peer','testpeerorg2.peer')" -f "InitLedger" \ 
  -w testnetwork -V v2.0 --initcc
~~~
Or you can build the networkconfig.yaml by yourself, for example:
~~~yaml
name: testnetwork
version: v2.0
node_image_tag: 2.3.3
ca_image_tag: 1.4.9
couch_image_tag: 0.4.22
channels:
  mychannel:
    consensus: solo
    peers:
    - peer.testpeerorg1
    - peer.testpeerorg2
    orderers:
    - orderer.testordererorg
    chaincodes:
    - name: mycc
      sequence: 1
nodes:
  orderer.testordererorg:
    name: orderer
    nodeport: 7050
    type: orderer
    org_id: testordererorg
    domain: testordererorg
    host: 127.0.0.1
    ssh_port: 22
  peer.testpeerorg1:
    name: peer
    nodeport: 7051
    type: peer
    org_id: testpeerorg1
    domain: testpeerorg1
    host: 127.0.0.1
    ssh_port: 22
  peer.testpeerorg2:
    name: peer
    nodeport: 8051
    type: peer
    org_id: testpeerorg2
    domain: testpeerorg2
    host: 127.0.0.1
    ssh_port: 22
chaincodes:
  mycc:
    path: $GOPATH/github.com/hyperledger/fabric-samples/chaincode/fabcar/go
    version: v1
    policy: OR('testpeerorg1.peer','testpeerorg2.peer')
    init_required: true
    init_func: InitLedger
~~~
And then, use the bootstrap command to generate files
~~~bash

~~~
Before starting the fabric network, you can modify files just generated, then using startup command to start the network.
~~~bash
fdt network startup -n testnetwork
~~~
If you want to stop the network, just using the shutdown command.
~~~bash
fdt network shutdown -n testnetwork
~~~
You can find more command in [scripts/localnet.sh](https://github.com/wjbbig/fabric-distributed-tool/blob/master/scripts/localnet.sh).

## Supported Features

---
* buildup and shutdown fabric network
* deploy or upgrade chaincode (only support chaincode-go)
* create or join channel
* extend new peer or orderer node
* start or stop the specified node
* register, enroll, revoke user with fabric-ca

## Licensing

---
Current fabric-distributed-tool code is released under [Apache License 2.0]( http://www.apache.org/licenses/)