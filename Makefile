all: dm_proto inv_proto dm_client fm inv_client dm_server inv_server dm_server_linux dm_client_linux inv_server_linux inv_client_linux
dm: dm_proto dm_client dm_server dm_server_linux dm_client_linux
fm: fm_proto fm_client fm_server
inv: inv_proto inv_client inv_server inv_server_linux inv_client_linux
dm_proto:
	(cd devicemanager; protoc -I protos -I ../ --go_out=plugins=grpc:${GOPATH}/src protos/device_manager.proto)
inv_proto:
	(cd inventory; protoc -I protos -I ../ --go_out=plugins=grpc:${GOPATH}/src protos/inventory.proto)
dm_server:
	go build -o build/dm_server devicemanager/server/server.go
dm_server_linux:
	GOOS=linux go build -o build/dm_server_linux devicemanager/server/server.go
dm_client:
	go build -o build/dm_client devicemanager/client/client.go
dm_client_linux:
	GOOS=linux go build -o build/dm_client_linux devicemanager/client/client.go
inv_server:
	go build -o build/inv_server inventory/server/server.go
inv_server_linux:
	GOOS=linux go build -o build/inv_server_linux inventory/server/server.go
inv_client:
	go build -o build/inv_client inventory/client/client.go
inv_client_linux:
	GOOS=linux go build -o build/inv_client_linux inventory/client/client.go
fm_proto:
	(cd fabricmanager; protoc -I protos -I ../ --go_out=plugins=grpc:${GOPATH}/src protos/fabric_manager.proto)
fm_server:
	go build -o build/fm_server fabricmanager/server/server.go
fm_client:
	go build -o build/fm_client fabricmanager/client/client.go
#container:
#	docker build -t michaelhenkel/remotexec . && docker push michaelhenkel/remotexec
