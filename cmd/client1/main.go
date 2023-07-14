package main

import (
	"context"
	"fmt"
	pb "go-rocket/example/proto/test"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	*pb.UnimplementedNewGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.TestRequest) (*pb.TestReply, error) {
	//md, _ := metadata.FromIncomingContext(ctx)
	fmt.Println(in.One)
	time.Sleep(800 * time.Millisecond)
	// 创建一个HelloReply消息，设置Message字段，然后直接返回。
	return &pb.TestReply{Message: "tests 1" + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//op := grpc.ForceServerCodec(codec.DefaultGRPCCodecs["application/proto"])
	// 实例化grpc服务端
	s := grpc.NewServer()

	// 注册Greeter服务
	pb.RegisterNewGreeterServer(s, new(server))

	// 往grpc服务端注册反射服务
	reflection.Register(s)

	// 启动grpc服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
