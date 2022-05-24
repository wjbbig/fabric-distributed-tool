This doc shows how to deploy a new chaincode or upgrade an existing chaincode.  
Once starting the fabric network, you can use the `deploycc` command to deploy new chaincode,
this command will install the chaincode on every peer in the network.

~~~bash
# if the version of network is v1.4.x, the 'ccpath' should be relative path based on gopath, 
# else it should be absolute path
fdt network deploycc -n testnetwork -c firstchannel --ccid mycc -p chaincode-go -v v1 --initcc \
  --policy "OR('testpeerorg1.peer','testpeerorg2.peer')" -f init -i "a" -i "100" -i "b" -i "100"
~~~

If the chaincode has been instantiated in the channel, the peer only needs to install the chaincode.
At this point, you should use `installcc` command.

~~~bash
fdt network installcc -n testnetwork -c firstchannel --ccid mycc -p peer.testpeerorg3
~~~

If you want to update the installed chaincode, using `upgradecc` command.

~~~bash
fdt network upgradecc -n testnetwork -c firstchannel -v v2 --initcc --policy "OR('testpeerorg1.peer')" \
  -f init -i "a" -i "100" -i "b" -i "100"
~~~

### ccaas

FDT supports ccaas mode if the version of peer is v2.4.x, but you have to prepare the connection.json, metadata.json and
the special packaged chaincode by yourself. Then using the `deploycc` command to deploy chaincode and setting
the `ccaas` flag
to true.  
Before starting the network, an env `CORE_CHAINCODE_MODE=dev` has to add to peer's docker-compose file. Then staring the
network without deploying any chaincode.

~~~ bash
fdt network startup -n testnetwork --startonly
~~~

Now, deploying and starting chaincode the way you like.

~~~bash
fdt network deploycc -n testnetwork -c firstchannel --ccid mycc -p /root/go/src/chaincode-go -v v1 \
--policy "OR('testpeerorg1.peer','testpeerorg2.peer')" --ccaas
~~~
