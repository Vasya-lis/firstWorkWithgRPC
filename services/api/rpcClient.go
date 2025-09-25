package api

import (
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	pb.SchedulerServiceClient
}

func NewSchedulerClient(conn *grpc.ClientConn) *SchedulerClient {
	return &SchedulerClient{
		SchedulerServiceClient: pb.NewSchedulerServiceClient(conn),
	}
}
