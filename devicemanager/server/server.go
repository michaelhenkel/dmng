package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

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
	mu sync.Mutex
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

func serialize(intf *dmPB.Interface) string {
	return fmt.Sprintf("%s", intf.Name)
}

func (d *deviceManagerServer) CreateInterface(stream dmPB.DeviceManager_CreateInterfaceServer) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		dbClient := database.NewDBClient()
		intf, err := stream.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		fmt.Printf("received intf: %s\n", intf.Name)
		if err := dbClient.ReadInterface(intf); err == nil {
			//st := status.New(codes.AlreadyExists, "Interface already exists")
			log.Println("object already exists")
			//return st.Err()
			result := &dmPB.Result{
				Received: true,
				Msg:      "failed",
			}
			if err := stream.Send(result); err != nil {
				log.Printf("send error %v", err)
			}
			continue
			//return err
		}
		intf.Device = &dmPB.Device{
			Name: *name,
		}
		if err := dbClient.CreateInterface(intf); err != nil {
			fmt.Println(err)
			return err
		}
		result := &dmPB.Result{
			Received: true,
			Msg:      "success",
		}
		if err := stream.Send(result); err != nil {
			log.Printf("send error %v", err)
		}
	}
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
