all: dm_proto inv_proto dm_client inv_client dm_server inv_server dm_server_linux dm_client_linux inv_server_linux inv_client_linux
dm: dm_proto dm_client dm_server dm_server_linux dm_client_linux
inv: inv_proto inv_client inv_server inv_server_linux inv_client_linux
dm_proto:
	(cd devicemanager; protoc -I protos -I ../ --go_out=plugins=grpc:${GOPATH}/src protos/device_manager.proto)
dm_server:
	(cd devicemanager/server; go build)
dm_server_linux:
	(cd devicemanager/server; GOOS=linux go build -o server_linux)
dm_client:
	(cd devicemanager/client; go build)
dm_client_linux:
	(cd devicemanager/client; GOOS=linux go build -o client_linux)
inv_proto:
	(cd inventory; protoc -I protos -I ../ --go_out=plugins=grpc:${GOPATH}/src protos/inventory.proto)
inv_server:
	(cd inventory/server; go build)
inv_server_linux:
	(cd inventory/server; GOOS=linux go build -o server_linux)
inv_client:
	(cd inventory/client; go build)
inv_client_linux:
	(cd inventory/client; GOOS=linux go build -o client_linux)
#container:
#	docker build -t michaelhenkel/remotexec . && docker push michaelhenkel/remotexec
