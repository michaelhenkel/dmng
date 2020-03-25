package main

import (
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/michaelhenkel/dmng/testing/test4/protos"
)

type simpleServer struct {
}

func (s *simpleServer) SimpleRPC(stream pb.SimpleService_SimpleRPCServer) error {
	log.Println("Started stream")
	for {
		in, err := stream.Recv()
		log.Println("Received value")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Println("Got " + in.Msg)
		log.Println("Sending result " + in.Msg)
		stream.Send(&pb.SimpleData{Msg: "Result " + in.GetMsg()})
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(20)
		log.Printf("sleeping for %d seconds\n", n)
		time.Sleep(time.Duration(n) * time.Second)

	}
}

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleServiceServer(grpcServer, &simpleServer{})

	l, err := net.Listen("tcp", ":16000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Listening on tcp://localhost:16000")
	grpcServer.Serve(l)
}
