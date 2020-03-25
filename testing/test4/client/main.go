package main

import (
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/michaelhenkel/dmng/testing/test4/protos"
)

func main() {
	conn, err := grpc.Dial("localhost:16000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer conn.Close()

	client := pb.NewSimpleServiceClient(conn)
	stream, err := client.SimpleRPC(context.Background())
	done := make(chan int)
	run := &runner{
		Stream:  stream,
		Send:    make(chan *pb.SimpleData),
		Receive: make(chan *pb.SimpleData),
	}
	go run.sender()
	go run.receiver()
	go func() {
		for i := 0; i < 20; i++ {
			msg := &pb.SimpleData{Msg: "msg " + strconv.Itoa(i)}
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(20)
			log.Printf("sleeping for %d seconds\n", n)
			log.Println("sending msg: " + msg.GetMsg())
			time.Sleep(time.Duration(n) * time.Second)
			run.Send <- msg
		}
	}()
	go func() {
		for {
			select {
			case msg := <-run.Receive:
				log.Println("received reply: " + msg.GetMsg())
			default:
			}
		}
	}()
	<-done
}

type runner struct {
	Stream  pb.SimpleService_SimpleRPCClient
	Send    chan *pb.SimpleData
	Receive chan *pb.SimpleData
}

func (r *runner) sender() {
	waitc := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-r.Send:
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
			resp, err := r.Stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			//log.Printf("new max %v received", resp)
			r.Receive <- resp

		}
	}()

	<-waitc
	r.Stream.CloseSend()
}
