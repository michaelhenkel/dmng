package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	fmPB "github.com/michaelhenkel/dmng/fabricmanager/protos"
)

var (
	port    = flag.Int("port", 10000, "The server port")
	name    = flag.String("name", "", "Device Name")
	request = flag.String("request", "request.yaml", "yaml")
)

func main() {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer conn.Close()
	client := fmPB.NewFabricManagerClient(conn)
	stream, err := client.RequestHandler(context.Background())
	var createRequest *fmPB.Request
	if *request != "" {
		requestYaml, err := ioutil.ReadFile(*request)
		if err != nil {
			log.Fatalln(err)
		}
		var fabricConfig fmPB.Fabric
		err = yaml.Unmarshal(requestYaml, &fabricConfig)
		if err != nil {
			fmt.Printf("Error parsing YAML file: %s\n", err)
		}
		createRequest = &fmPB.Request{
			Request: &fmPB.Request_Create{
				Create: &fmPB.Create{
					CreateRequest: &fmPB.Create_Fabric{
						Fabric: &fabricConfig,
					},
				},
			},
		}
		log.Println("sending request: ", createRequest)
		//run.Send <- request
	}
	ctx := stream.Context()
	done := make(chan bool)
	go func(createRequest *fmPB.Request) {
		if err := stream.Send(createRequest); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}(createRequest)
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("received result %+v\n", resp)
		}
	}()
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
}

type runner struct {
	Stream  fmPB.FabricManager_RequestHandlerClient
	Send    chan *fmPB.Request
	Receive chan *fmPB.Result
	Done    chan bool
}

func (r *runner) sender() {
	done := make(chan bool)
	go func() {
		for {
			select {
			case msg := <-r.Send:
				r.Stream.Send(msg)
				if err := r.Stream.CloseSend(); err != nil {
					log.Println(err)
				}
			default:
			}
		}

	}()
	<-done
	r.Stream.CloseSend()
}

func (r *runner) receiver() {
	done := make(chan bool)
	go func() {
		for {
			resp, err := r.Stream.Recv()
			if err == io.EOF {
				r.Stream.CloseSend()

				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Println("received reply: " + resp.GetMsg())
			if err := r.Stream.CloseSend(); err != nil {
				log.Println(err)
			}
			//close(done)
		}
	}()
	<-done
	r.Stream.CloseSend()
}

func (r *runner) senderreceiver(request *fmPB.Request) {
	done := make(chan bool)
	go func() {
		for {
			resp, err := r.Stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				r.Stream.CloseSend()
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Println("received reply: " + resp.GetMsg())
			if err := r.Stream.CloseSend(); err != nil {
				log.Println(err)
			}
		}
	}()
	go func() {
		<-r.Stream.Context().Done()

		if err := r.Stream.Context().Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()
	r.Stream.Send(request)
	if err := r.Stream.CloseSend(); err != nil {
		log.Println(err)
	}
	<-done
}
