package main

import (
	pb1 "go-rocket/example/proto/hello"
	pb2 "go-rocket/example/proto/test"
	"go-rocket/mash"
	"go-rocket/service"
)

func main() {
	//fmt.Println(reflect.TypeOf(&pb.HelloRequest{}).Elem().PkgPath())
	// mash := mash.NewHttpMash()
	// mash.BuliderRouter(
	// 	service.BuildRegisterMessage(&pb.HelloRequest{}, &pb.HelloReply{}),
	// 	service.BuilderRegCenter(service.NewLocalCenterNoHost([]*service.RouterInfo{
	// 		{
	// 			Path:       "proto/Greeter/SayHello",
	// 			Host:       "127.0.0.1:50051",
	// 			InMessage:  &pb.HelloRequest{},
	// 			OutMessage: &pb.HelloReply{},
	// 		},
	// 	})),
	// )
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
	// err := mash.ListenWithPort(":9000")
	// if err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
	mash.NewMashContainer(":9000", ":9008",
		service.BuildRegisterMessage(&pb1.HelloRequest{}, &pb1.HelloReply{}, &pb2.TestRequest{}, &pb2.TestReply{}),
		service.BuilderRegCenter(service.NewLocalCenterNoHost([]*service.RouterInfo{
			{
				Path:       "proto/Greeter/SayHello",
				Host:       "127.0.0.1:50051",
				InMessage:  &pb1.HelloRequest{},
				OutMessage: &pb1.HelloReply{},
			}, {
				Path:       "proto/NewGreeter/SayHello",
				Host:       "127.0.0.1:50052",
				InMessage:  &pb2.TestRequest{},
				OutMessage: &pb2.TestReply{},
			},
		}))).Listen()
}
