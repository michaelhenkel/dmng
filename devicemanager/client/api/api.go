package api

import (
	"context"
	"log"
	"time"

	pbDM "github.com/michaelhenkel/dmng/devicemanager/protos"
	"google.golang.org/grpc"
)

func newClient(server_address string) (pbDM.DeviceManagerClient, context.Context, *grpc.ClientConn, context.CancelFunc) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(server_address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pbDM.NewDeviceManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	return c, ctx, conn, cancel
}

type Connection struct {
	ServerAddress string
}

func (c *Connection) CreateInterface(intf *pbDM.Interface) (pbDM.DeviceManager_CreateInterfaceClient, error) {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress)
	defer conn.Close()
	defer cancel()
	return pbDMClient.CreateInterface(ctx, intf)
}

func (c *Connection) ReadInterface(intf *pbDM.Interface) (*pbDM.Result, error) {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress)
	defer conn.Close()
	defer cancel()
	return pbDMClient.ReadInterface(ctx, intf)
}

func (c *Connection) DeleteInterface(intf *pbDM.Interface) (*pbDM.Result, error) {
	pbDMClient, ctx, conn, cancel := newClient(c.ServerAddress)
	defer conn.Close()
	defer cancel()
	return pbDMClient.DeleteInterface(ctx, intf)
}

/*
func (e *Executor) GetFileContent(filePath string) (*string, error) {
	socket := e.Socket
	c, ctx, conn, cancel := newClient(&socket)
	defer conn.Close()
	defer cancel()
	fileResult, err := c.GetFileContent(ctx, &protos.FilePath{Path: filePath})
	if err != nil {
		return nil, err
	}
	return &fileResult.Result, nil
}
*/
