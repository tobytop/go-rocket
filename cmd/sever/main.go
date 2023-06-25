package main

import (
	pb "go-rocket/example/proto/hello"
	"go-rocket/mash"
	"go-rocket/service"

	"log"
)

func main() {
	//fmt.Println(reflect.TypeOf(&pb.HelloRequest{}).Elem().PkgPath())
	mash := mash.NewMash()
	mash.BuliderRouter(
		service.BuilderBalance(service.WeightRobin),
		service.BuildRegMessage(&pb.HelloRequest{}, &pb.HelloReply{}),
		service.BuilderRegCenter(service.NewLocalCenter(map[string]int{
			"127.0.0.1:50051": 1,
			"127.0.0.1:50052": 1,
		}, []*service.RouterInfo{
			{
				PackageName: "proto",
				ServiceName: "Greeter",
				Method:      "SayHello",
				InMessage:   &pb.HelloRequest{},
				OutMessage:  &pb.HelloReply{},
			}})))
	err := mash.ListenWithPort(":9000")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
