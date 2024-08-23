package main

import (
	"context"
	"log"
	"net"
	"pro/service"

	pb "pro/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedWeightSetServiceServer
} //服务对象

// 实现服务的接口 在proto中定义的所有服务都是接口
func (s *server) Set(ctx context.Context, in *pb.WeightConfig) (*emptypb.Empty, error) {
	var numNode = len(in.Weigtht.Rows)
	var scheduleDesion = make([][]int, numNode)
	for i := range scheduleDesion {
		scheduleDesion[i] = make([]int, numNode) // 初始化每个切片的长度为numNode
	}
	for i := 0; i < numNode; i++ {
		for j := 0; j < numNode; j++ {
			scheduleDesion[i][j] = int(in.Weigtht.Rows[i].Values[j])
		}
	}
	err := service.WeightSet(in.Service, scheduleDesion, in.Namespace)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer() //起一个服务

	pb.RegisterWeightSetServiceServer(s, &server{})
	// 注册反射服务 这个服务是CLI使用的 跟服务本身没有关系
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
