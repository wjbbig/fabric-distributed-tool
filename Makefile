build:
	go build -o fdt cmd/root_cmd.go

clean:
	rm -rf /opt/fdt

generate:
	./fdt generate bootstrap -p peer.testpeerorg:7051@root@101.91.124.43:22:Blockchain@123 -o orderer.testordererorg:7050@root@101.91.231.36:22:Blockchain@123 -s solo -d /opt/fdt -c mychannel -v 1.4

startup:
	./fdt startup --datadir /opt/fdt