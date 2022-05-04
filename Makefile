build:
	go build -o fdt cmd/root_cmd.go

clean:
	rm -rf fdtdata/*
	rm -rf $(HOME)/.fdt/network.json

