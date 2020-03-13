package server

import (
	"flag"
	"fmt"
	"log"
	"net"

	invPb "github.com/michaelhenkel/dmng/inventory/protos"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10001, "The server port")
)

type inventoryServer struct {
	invPb.UnimplementedInventoryServer
}

func newServer() *inventoryServer {
	s := &inventoryServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	invPb.RegisterInventoryServer(grpcServer, newServer())
	grpcServer.Serve(lis)

}
