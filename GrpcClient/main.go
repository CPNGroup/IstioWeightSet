package main

import (
	"context"
	"log"
	"time"

	pb "pro/pb"

	"google.golang.org/grpc"
)

const (
	address = "10.200.1.43:30031"
)

func main() {

	//建立链接
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewWeightSetServiceClient(conn)

	// 1秒的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 先定一个调度决策变量
	var numNode = 3
	var scheduleDesion = make([][]int32, numNode)
	for i := range scheduleDesion {
		scheduleDesion[i] = make([]int32, numNode) // 初始化每个切片的长度为numNode
	}
	// 赋值
	// nuc1的卸载决策
	scheduleDesion[0][0] = 50
	scheduleDesion[0][1] = 50
	scheduleDesion[0][2] = 0
	// nuc2的卸载决策
	scheduleDesion[1][0] = 100
	scheduleDesion[1][1] = 0
	scheduleDesion[1][2] = 0
	// nuc3的卸载决策
	scheduleDesion[2][0] = 0
	scheduleDesion[2][1] = 0
	scheduleDesion[2][2] = 100

	// 根据scheduleDesion转换为pb.Matrix
	matrix := &pb.Matrix{}
	for i := 0; i < numNode; i++ {
		row := &pb.Row{Values: scheduleDesion[i]}
		matrix.Rows = append(matrix.Rows, row)
	}

	// 定义请求参数
	var weightConfig = pb.WeightConfig{Namespace: "li", Service: "server", Weigtht: matrix}
	_, err = c.Set(ctx, &weightConfig)
	if err != nil {
		log.Fatalf("could not set: %v", err)
	}
	log.Printf("Success")
}
