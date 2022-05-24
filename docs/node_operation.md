This doc shows how to use fdt to operate fabric node and channel.
Before using `generate --extend` command to generate node keypairs, you should add nodes into the networkconfig.yaml.
~~~yaml
#...
nodes:
  #...
  # new nodes
  peer1.testpeerorg2:
    name: peer
    nodeport: 8051
    type: peer1
    org_id: testpeerorg2
    domain: testpeerorg2
    host: 192.168.1.2
    ssh_port: 22
    dest: /data/fabric-distributed-tool/fdtdata
  peer1.testpeerorg3:
    name: peer1
    nodeport: 9051
    type: peer
    org_id: testpeerorg3
    domain: testpeerorg3
    host: 192.168.1.2
    ssh_port: 22
    dest: /data/fabric-distributed-tool/fdtdata
  orderer1.testordererorg:
    name: orderer1
    nodeport: 7050
    type: orderer
    org_id: testordererorg
    domain: testordererorg
    host: 192.168.1.2
    ssh_port: 22
    dest: /root/fabric-distributed-tool/fdtdata
#...
~~~
then using the `generate --extend` command.
~~~bash
fdt network generate --extend -n testwork -p peer1.testpeerorg2 -p peer1.testpeerorg3 -o orderer1.testordererorg
~~~
The above command will not start the node, you have to use `startnode` command to start the node you want to,
~~~bash
fdt network startnode -n testnetwork -p peer1.testpeerorg3 -p peer1.testpeerorg2
~~~
also, you can use `stopnode` command to stop a node.
~~~bash
fdt network stopnode -n testnetwork -p peer1.testpeerorg3
~~~
Once starting a node, you can let the node join an existing channel, if the organization of the node has already joined the channel, you should use `existorgjoin` command,
~~~bash
fdt network existorgjoin -n testnetwork -c firstchannel -p peer1.testpeerorg2
~~~
else using the `neworgjoin` command to join the new channel.
~~~bash
fdt network existorgjoin -n testnetwork -c firstchannel -p peer1.testpeerorg3
~~~
Except for joining the channel, you can also use `createchannel` command to create a new channel.
~~~bash
fdt network createchannel -n testnetwork -c newchannel -p peer1.testpeerorg3 -p peer1.testpeerorg2
~~~
