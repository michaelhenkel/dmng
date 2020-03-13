package server

import (
	"flag"
	"fmt"
	"log"
	"net"

	dmPb "github.com/michaelhenkel/dmng/devicemanager/protos"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type deviceManagerServer struct {
	dmPb.UnimplementedDeviceManagerServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	dmPb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)

}
