package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
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
	stream  dmPB.DeviceManager_RequestHandlerServer
	clients map[string][]chan *dmPB.Request
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

func (d *deviceManagerServer) RequestHandler(stream dmPB.DeviceManager_RequestHandlerServer) error {
	log.Println("Started stream")
	for {
		request, err := stream.Recv()
		log.Println("Received value")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("%+v\n", request.GetCreate())
		result := &dmPB.Result{
			Received: true,
			Msg:      msg,
			Applied:  false,
		}
		if err := stream.Send(result); err != nil {
			log.Printf("send error %v", err)
		}
		/*
			createObj := request.GetCreate()
			if createObj != nil {
				if err := d.create(createObj); err != nil {
					return err
				}
			}
		*/
	}
}

func (d *deviceManagerServer) create(createObj *dmPB.Create) error {
	if createObj.GetInterfaces() != nil {
		if err := d.createInterface(createObj.GetInterfaces()); err != nil {
			return err
		}
	}
	return nil
}

func (d *deviceManagerServer) createInterface(intfList *dmPB.Interfaces) error {
	dbClient := database.NewDBClient()
	result := &dmPB.Result{}
	for _, intf := range intfList.GetInterface() {
		if err := dbClient.ReadInterface(intf); err != nil {
			intf.Version = 1
			result.Msg = "creating new object"
		} else {
			intf.Version++
			result.Msg = "object already exists, updating"
		}
		result := &dmPB.Result{
			Received: true,
			Msg:      "create interface failed",
			Success:  false,
			Applied:  false,
		}
		if err := d.stream.Send(result); err != nil {
			log.Printf("send error %v", err)
		}
		intf.Device = &dmPB.Device{
			Name: *name,
		}

		// Here the device driver API will be invoced

		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(5)
		log.Printf("sleeping for %d seconds\n", n)
		time.Sleep(time.Duration(n) * time.Second)

		// End of device driver invocation

		if err := dbClient.CreateInterface(intf); err != nil {
			fmt.Println(err)

			result := &dmPB.Result{
				Received: true,
				Msg:      "create interface failed",
				Success:  false,
				Applied:  false,
			}
			if err := d.stream.Send(result); err != nil {
				log.Printf("send error %v", err)
			}

			//continue
		}

		result = &dmPB.Result{
			Received: true,
			Msg:      "interface created",
			Success:  true,
			Applied:  true,
		}
		if err := d.stream.Send(result); err != nil {
			log.Printf("send error %v", err)
		}
	}

	return nil
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
	//opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaep))
	//opts = append(opts, grpc.KeepaliveParams(kasp))
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
