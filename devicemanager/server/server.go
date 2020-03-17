package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

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
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		result := &dmPB.Result{
			Received: true,
			Msg:      "request received",
			Applied:  false,
		}
		if err := stream.Send(result); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("received intf: %s\n", intf.Name)
		if err := dbClient.ReadInterface(intf); err == nil {
			log.Println("object already exists")
			result := &dmPB.Result{
				Received: true,
				Msg:      "object already exists",
				Success:  false,
				Applied:  false,
			}
			if err := stream.Send(result); err != nil {
				log.Printf("send error %v", err)
			}
			continue
		}
		intf.Device = &dmPB.Device{
			Name: *name,
		}
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(20)
		log.Printf("sleeping for %d seconds\n", n)
		time.Sleep(time.Duration(n) * time.Second)
		if err := dbClient.CreateInterface(intf); err != nil {
			fmt.Println(err)
			result := &dmPB.Result{
				Received: true,
				Msg:      "create interface failed",
				Success:  false,
				Applied:  false,
			}
			if err := stream.Send(result); err != nil {
				log.Printf("send error %v", err)
			}
			continue
		}
		result = &dmPB.Result{
			Received: true,
			Msg:      "interface created",
			Success:  true,
			Applied:  true,
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
