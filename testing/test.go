package test

import (
	"context"
	"fmt"

	pb "github.com/michaelhenkel/dmng/testing/protos"
	"google.golang.org/grpc"
)

func main() {

	var conn *grpc.ClientConn
	fabricManagerClient := pb.NewFabricManagerClient(conn)

	clos := &pb.CLOS{
		Leaves: []*pb.Device{{
			Name: "device1",
			Roles: []pb.Role{
				pb.Role_ERB,
				pb.Role_LEAF,
			},
		}, {
			Name: "device2",
			Roles: []pb.Role{
				pb.Role_ERB,
				pb.Role_LEAF,
			},
		}},
		Spines: []*pb.Device{{
			Name: "device3",
			Roles: []pb.Role{
				pb.Role_CRB,
				pb.Role_SPINE,
			},
		}},
	}
	fabric := &pb.Fabric{
		Name: "fab1",
		Topology: &pb.Fabric_Clos{
			Clos: clos,
		},
	}

	stream, _ := fabricManagerClient.CreateFabric(context.Background())
	stream.Send(fabric)
	result, _ := stream.Recv()

	fmt.Println(result)
}
