package api

import (
	"context"
	"io"
	"log"
	"time"

	pbDM "github.com/michaelhenkel/dmng/devicemanager/protos"
	"google.golang.org/grpc"
)

func newClient(server_address string, timeout int) (pbDM.DeviceManagerClient, context.Context, *grpc.ClientConn, context.CancelFunc) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server_address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pbDM.NewDeviceManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	log.Printf("timeout: %d\n", timeout)
	return c, ctx, conn, cancel
}

type Connection struct {
	ServerAddress string
}

func (c *Connection) CreateInterface(intfList []*pbDM.Interface, timeout int) error {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress, timeout)
	defer conn.Close()
	defer cancel()
	stream, err := pbDMClient.CreateInterface(ctx)
	if err != nil {
		return err
	}
	done := make(chan bool)
	go func() {
		for _, intf := range intfList {
			if err := stream.Send(intf); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("%+v\n", result)
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
	return nil
}

func (c *Connection) ReadInterface(intf *pbDM.Interface, timeout int) (*pbDM.Result, error) {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress, timeout)
	defer conn.Close()
	defer cancel()
	return pbDMClient.ReadInterface(ctx, intf)
}

func (c *Connection) DeleteInterface(intf *pbDM.Interface, timeout int) (*pbDM.Result, error) {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress, timeout)
	defer conn.Close()
	defer cancel()
	return pbDMClient.DeleteInterface(ctx, intf)
}
