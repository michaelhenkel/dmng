package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/michaelhenkel/dmng/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
)

var (
	port = flag.Int("port", 10000, "The server port")
	name = flag.String("name", "", "Device Name")
)

type deviceManagerServer struct {
	dmPB.UnimplementedDeviceManagerServer
}

func newServer() *deviceManagerServer {
	s := &deviceManagerServer{}
	return s
}

func (d *deviceManagerServer) CreateDevice(ctx context.Context, device *dmPB.Device) (*dmPB.Device, error) {
	return device, nil
}

func (d *deviceManagerServer) ReadDevice(ctx context.Context, device *dmPB.Device) (*dmPB.Device, error) {
	return device, nil
}

func (d *deviceManagerServer) CreateInterface(ctx context.Context, intf *dmPB.Interface) (*dmPB.Result, error) {
	dbClient := database.NewDBClient()
	result := &dmPB.Result{}
	if err := dbClient.ReadInterface(intf); err == nil {
		result.Msg = "failed"
		st := status.New(codes.AlreadyExists, "Interface already exists")
		return result, st.Err()
	}
	intf.Device = &dmPB.Device{
		Name: *name,
	}
	if err := dbClient.CreateInterface(intf); err != nil {
		fmt.Println(err)
		result.Msg = "failed"
	}
	result.Msg = "success"
	return result, nil
}

func (d *deviceManagerServer) ReadInterface(ctx context.Context, intf *dmPB.Interface) (*dmPB.Result, error) {
	dbClient := database.NewDBClient()
	result := &dmPB.Result{}
	if err := dbClient.ReadInterface(intf); err != nil {
		st := status.New(codes.NotFound, "Interface does not exist")
		result.Msg = "failed"
		return result, st.Err()
	}
	result.Msg = "success"
	return result, nil
}

func (d *deviceManagerServer) DeleteInterface(ctx context.Context, intf *dmPB.Interface) (*dmPB.Result, error) {
	dbClient := database.NewDBClient()
	result := &dmPB.Result{}
	if err := dbClient.ReadInterface(intf); err != nil {
		st := status.New(codes.NotFound, "Interface does not exist")
		result.Msg = "failed"
		return result, st.Err()
	}
	if err := dbClient.DeleteInterface(intf); err != nil {
		return result, err
	}
	result.Msg = "success"
	return result, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	dev := &dmPB.Device{
		Name: *name,
	}
	dbClient := database.NewDBClient()
	if err := dbClient.ReadDevice(dev); err == err.(*database.ObjectNotFound) {
		fmt.Println(err)
		if err = dbClient.CreateDevice(dev); err != nil {
			fmt.Println(err)
		}
	}
	dmPB.RegisterDeviceManagerServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
