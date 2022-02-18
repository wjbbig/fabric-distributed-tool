build-arm:
	GOOS=darwin GOARCH=arm64 go build -o fdt cmd/root_cmd.go

build-amd:
	GOOS=linux GOARCH=amd64 go build -o fdt cmd/root_cmd.go
clean:
	rm -rf /opt/fdt

