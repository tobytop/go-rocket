package main

import (
	"context"
	pb "go-rocket/example/protos/hello"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	*pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// 创建一个HelloReply消息，设置Message字段，然后直接返回。
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 实例化grpc服务端
	s := grpc.NewServer()

	// 注册Greeter服务
	pb.RegisterGreeterServer(s, new(server))

	// 往grpc服务端注册反射服务
	reflection.Register(s)

	// 启动grpc服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
