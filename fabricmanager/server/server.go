package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
	fmPB "github.com/michaelhenkel/dmng/fabricmanager/protos"
)

var (
	port             = flag.Int("port", 10000, "The server port")
	dmAgentMap       = make(map[string]*dmAgent)
	fmRun            *fmRunner
	fmReceiveChannel = make(chan *fmPB.Result)
)

type deviceManagerServer struct {
	dmPB.UnimplementedDeviceManagerServer
}

type fabricManagerServer struct {
	fmPB.UnimplementedFabricManagerServer
}

func newDMServer() *deviceManagerServer {
	s := &deviceManagerServer{}
	return s
}

func newFMServer() *fabricManagerServer {
	s := &fabricManagerServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on localhost:%d\n", *port)
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	dmPB.RegisterDeviceManagerServer(grpcServer, newDMServer())
	fmPB.RegisterFabricManagerServer(grpcServer, newFMServer())
	grpcServer.Serve(lis)
}

type dmAgent struct {
	Name    string
	Stream  dmPB.DeviceManager_RequestHandlerServer
	Request chan *dmPB.Message
	Result  chan *dmPB.Message
}

func (f *fabricManagerServer) RequestHandler(stream fmPB.FabricManager_RequestHandlerServer) error {
	log.Println("Started FM streaming server")
	ctx := stream.Context()
	done := make(chan bool)

	go func() {
		for {
			select {
			case res := <-fmReceiveChannel:
				log.Println("received on channel", res)
				if err := stream.Send(res); err != nil {
					log.Println(err)
				}
			default:
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		createInterface(req)
		if err := stream.Send(&fmPB.Result{Msg: "blabla"}); err != nil {
			log.Println(err)
		}

	}
	<-done
	return nil
}

type fmRunner struct {
	Stream  fmPB.FabricManager_RequestHandlerServer
	Send    chan *fmPB.Result
	Receive chan *fmPB.Request
	Done    chan bool
}

func createInterface(req *fmPB.Request) {
	create := req.GetCreate()
	fabric := create.GetFabric()
	deviceList := fabric.GetDevices()
	for _, device := range deviceList {
		if _, ok := dmAgentMap[device]; !ok {
			log.Fatalf("Agent %s not connected", device)
		}
	}
	for _, device := range deviceList {
		dmAgent := dmAgentMap[device]
		message := &dmPB.Message{
			Message: &dmPB.Message_Request{
				Request: &dmPB.Request{
					Request: &dmPB.Request_Create{
						Create: &dmPB.Create{
							CreateRequest: &dmPB.Create_Interfaces{
								Interfaces: &dmPB.Interfaces{
									Interface: []*dmPB.Interface{{
										Name: "eth0",
										Ipv4: "1.1.1.1",
									}},
								},
							},
						},
					},
				},
			},
		}
		dmAgent.Request <- message
	}
}

func (r *fmRunner) senderreciever() {
	done := make(chan bool)
	go func() {
		for {
			select {
			case msg := <-r.Send:
				r.Stream.Send(msg)
			default:
			}
		}
	}()
	go func() {
		for {
			resp, err := r.Stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			//request := in.GetRequest()
			create := resp.GetCreate()
			fabric := create.GetFabric()
			deviceList := fabric.GetDevices()
			for _, device := range deviceList {
				if _, ok := dmAgentMap[device]; !ok {
					log.Fatalf("Agent %s not connected", device)
				}
			}
			for _, device := range deviceList {
				dmAgent := dmAgentMap[device]
				message := &dmPB.Message{
					Message: &dmPB.Message_Request{
						Request: &dmPB.Request{
							Request: &dmPB.Request_Create{
								Create: &dmPB.Create{
									CreateRequest: &dmPB.Create_Interfaces{
										Interfaces: &dmPB.Interfaces{
											Interface: []*dmPB.Interface{{
												Name: "eth0",
												Ipv4: "1.1.1.1",
											}},
										},
									},
								},
							},
						},
					},
				}
				dmAgent.Request <- message
			}
			//r.Receive <- resp
		}
	}()
	go func() {
		<-r.Stream.Context().Done()
		if err := r.Stream.Context().Err(); err != nil {
			log.Println(err)
		}
		close(r.Done)
	}()
	<-done
}

func (r *fmRunner) sender() {
	for {
		select {
		case msg := <-r.Send:
			r.Stream.Send(msg)
		default:
		}
	}

	//r.Stream.CloseSend()
}

func (r *fmRunner) receiver() {
	for {
		resp, err := r.Stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("can not receive %v", err)
		}
		//request := in.GetRequest()
		create := resp.GetCreate()
		fabric := create.GetFabric()
		deviceList := fabric.GetDevices()
		for _, device := range deviceList {
			if _, ok := dmAgentMap[device]; !ok {
				log.Fatalf("Agent %s not connected", device)
			}
		}
		for _, device := range deviceList {
			dmAgent := dmAgentMap[device]
			message := &dmPB.Message{
				Message: &dmPB.Message_Request{
					Request: &dmPB.Request{
						Request: &dmPB.Request_Create{
							Create: &dmPB.Create{
								CreateRequest: &dmPB.Create_Interfaces{
									Interfaces: &dmPB.Interfaces{
										Interface: []*dmPB.Interface{{
											Name: "eth0",
											Ipv4: "1.1.1.1",
										}},
									},
								},
							},
						},
					},
				},
			}
			dmAgent.Request <- message
		}
		//r.Receive <- resp
	}
}

func (d *deviceManagerServer) RequestHandler(stream dmPB.DeviceManager_RequestHandlerServer) error {
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		agentNameSlice := md.Get("client")
		log.Printf("added new agent %s to map", agentNameSlice[0])
		dma := &dmAgent{
			Stream:  stream,
			Request: make(chan *dmPB.Message),
			Result:  make(chan *dmPB.Message),
		}
		dmAgentMap[agentNameSlice[0]] = dma
		done := make(chan int)
		go dma.sender()
		go dma.receiver()
		<-done
	}
	return nil
}

func (r *dmAgent) sender() {
	log.Println("starting sender")
	for {
		select {
		case msg := <-r.Request:
			//log.Println("got message to be sent")
			if err := r.Stream.Send(msg); err != nil {
				log.Fatalf("couldn't send: %+v\n", err)
			}

		default:
		}
	}
}

func (r *dmAgent) receiver() {
	log.Println("starting receiver")
	for {
		msg, err := r.Stream.Recv()
		if err == io.EOF {
			//close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("can not receive %v", err)
		}
		result := msg.GetResult()
		request := msg.GetRequest()

		if request != nil {
			//log.Printf("received request: %+v\n", request)
			if connect := request.GetConnect(); connect != nil {
				//log.Printf("received connect request: %+v\n", connect)
				msg := &dmPB.Message{
					Message: &dmPB.Message_Result{
						Result: &dmPB.Result{
							Received: true,
							Msg:      "connection request successful for " + connect.GetClient(),
						},
					},
				}
				//log.Println("sending ack")
				r.Request <- msg
			}
		}

		if result != nil {
			dmRes := &fmPB.Result{
				Received: result.Received,
				Applied:  result.Applied,
				Msg:      result.Msg,
				Success:  result.Success,
			}
			//fmRun.Send <- dmRes
			log.Println("sending to receive chan")
			fmReceiveChannel <- dmRes
			//log.Printf("received result: %+v\n", result)
		}
	}
}
