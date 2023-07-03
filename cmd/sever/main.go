package main

import (
	pb "go-rocket/example/proto/hello"
	"go-rocket/mash"
	"go-rocket/service"

	"log"
)

func main() {
	//fmt.Println(reflect.TypeOf(&pb.HelloRequest{}).Elem().PkgPath())
	mash := mash.NewHttpMash()
	mash.BuliderRouter(
		service.BuildRegisterMessage(&pb.HelloRequest{}, &pb.HelloReply{}),
		service.BuilderRegCenter(service.NewLocalCenterNoHost([]*service.RouterInfo{
			{
				Path:       "proto/Greeter/SayHello",
				Host:       "127.0.0.1:50051",
				InMessage:  &pb.HelloRequest{},
				OutMessage: &pb.HelloReply{},
			},
		})),
	)
	// mash.BuliderRouter(
	// 	service.BuilderBalance(service.WeightRobin),
	// 	service.BuildRegisterMessage(&pb.HelloRequest{}, &pb.HelloReply{}),
	// 	service.BuilderRegCenter(service.NewLocalCenter(map[string]int{
	// 		"127.0.0.1:50051": 1,
	// 		"127.0.0.1:50052": 1,
	// 	}, []*service.RouterInfo{
	// 		{
	// 			Path:       "proto/Greeter/SayHello",
	// 			InMessage:  &pb.HelloRequest{},
	// 			OutMessage: &pb.HelloReply{},
	// 		}})))
	err := mash.ListenWithPort(":9000")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
