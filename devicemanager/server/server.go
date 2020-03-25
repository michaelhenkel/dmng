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

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
)

var (
	port     = flag.Int("port", 10000, "The server port")
	name     = flag.String("name", "", "Device Name")
	outbound = flag.Bool("outbound", true, "")
	server   = flag.String("server", "localhost:10000", "")
)

type deviceManagerServer struct {
	dmPB.UnimplementedDeviceManagerServer
}

func newServer() *deviceManagerServer {
	s := &deviceManagerServer{}
	return s
}

type runner struct {
	Stream  dmPB.DeviceManager_RequestHandlerClient
	Request chan *dmPB.Message
	Result  chan *dmPB.Message
}

func main() {
	flag.Parse()
	if !*outbound {
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		dmPB.RegisterDeviceManagerServer(grpcServer, newServer())
		grpcServer.Serve(lis)
	} else {
		log.Println("initiating outbound connection to", *server)
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(*server, opts...)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		if err != nil {
			log.Fatalf("failed to connect: %s", err)
		}
		defer conn.Close()

		md := metadata.Pairs(
			"client", *name,
		)

		client := dmPB.NewDeviceManagerClient(conn)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		stream, err := client.RequestHandler(ctx)
		if err != nil {
			log.Fatal(err)
		}
		run := &runner{
			Stream:  stream,
			Request: make(chan *dmPB.Message),
			Result:  make(chan *dmPB.Message),
		}
		done := make(chan int)
		go run.sender()
		go run.receiver()
		connectMessage := &dmPB.Message{
			Message: &dmPB.Message_Request{
				Request: &dmPB.Request{
					Request: &dmPB.Request_Connect{
						&dmPB.Connect{
							Client: *name,
						},
					},
				},
			},
		}
		run.Request <- connectMessage
		<-done
	}
}

func (r *runner) sender() {
	waitc := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-r.Request:
				r.Stream.Send(msg)
			default:
			}
		}
	}()
	<-waitc
	r.Stream.CloseSend()
}

func (r *runner) receiver() {
	waitc := make(chan struct{})
	go func() {
		for {
			msg, err := r.Stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			result := msg.GetResult()
			request := msg.GetRequest()

			if request != nil {
				log.Printf("received request: %+v\n", request)
				create := request.GetCreate()
				if create != nil {
					interfaces := create.GetInterfaces()
					if interfaces != nil {
						for _, intf := range interfaces.GetInterface() {
							rand.Seed(time.Now().UnixNano())
							n := rand.Intn(5)
							log.Printf("Creating interface %+v in %d sec\n", intf, n)
							time.Sleep(time.Duration(n) * time.Second)
							log.Printf("Interface %+v created\n", intf)
							result := &dmPB.Message{
								Message: &dmPB.Message_Result{
									Result: &dmPB.Result{
										Received: true,
										Applied:  true,
										Msg:      "created interface " + intf.GetName(),
										Success:  true,
									},
								},
							}
							r.Request <- result
						}
					}
				}
			}
			if result != nil {
				log.Printf("received result: %+v\n", result)
			}
		}
	}()

	<-waitc
	r.Stream.CloseSend()
}

func (d *deviceManagerServer) RequestHandler(stream dmPB.DeviceManager_RequestHandlerServer) error {
	in, err := stream.Recv()
	log.Println("Received value")
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	in.GetRequest().Descriptor()
	return nil
}
