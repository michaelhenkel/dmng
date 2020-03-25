package api

import (
	"context"
	"io"
	"log"

	pbDM "github.com/michaelhenkel/dmng/devicemanager/protos"
	"google.golang.org/grpc"
)

func (d *DMClient) NewClient() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(d.Address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer conn.Close()

	client := pbDM.NewDeviceManagerClient(conn)
	stream, err := client.RequestHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	d.stream = stream
	//d.Request = make(chan *pbDM.Request)
	//d.Result = make(chan *pbDM.Result)

	done := make(chan int)
	go d.sender()
	go d.receiver()
	<-done
}

type DMClient struct {
	Address string
	stream  pbDM.DeviceManager_RequestHandlerClient
	Request chan *pbDM.Request
	Result  chan *pbDM.Result
}

func (d *DMClient) sender() {
	waitc := make(chan struct{})
	log.Println("Sender is running")
	//go func() {
	for {
		select {
		case msg := <-d.Request:
			log.Println("received msg to be sent")
			d.stream.Send(msg)
		default:
		}
	}
	//}()
	<-waitc
	d.stream.CloseSend()
	log.Println("Sender is stopped")
}

func (d *DMClient) receiver() {
	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := d.stream.Recv()
			log.Println("received result")
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			d.Result <- resp
		}
	}()
	<-waitc
	d.stream.CloseSend()
}

func (d *DMClient) SendRquest(request *pbDM.Request) {
	log.Println("received request to be sent")
	d.Request <- request
	log.Println("sent request")
}
