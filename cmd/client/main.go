package main

import (
	"context"
	pb1 "go-rocket/example/proto/hello"
	pb2 "go-rocket/example/proto/test"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type server struct {
	*pb1.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb1.HelloRequest) (*pb1.HelloReply, error) {
	conn, err := grpc.Dial(":9008", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb2.NewNewGreeterClient(conn)
	// 创建一个HelloReply消息，设置Message字段，然后直接返回。
	md := metadata.MD{}
	log.Println(md)
	reply, err := c.SayHello(metadata.NewOutgoingContext(ctx, md), &pb2.TestRequest{Name: in.Name})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb1.HelloReply{Message: reply.Message}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//op := grpc.ForceServerCodec(codec.DefaultGRPCCodecs["application/proto"])
	// 实例化grpc服务端
	s := grpc.NewServer()

	// 注册Greeter服务
	pb1.RegisterGreeterServer(s, new(server))

	// 往grpc服务端注册反射服务
	reflection.Register(s)

	// 启动grpc服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
