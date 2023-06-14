package main

import (
	pb "go-rocket/example/proto/hello"
	"go-rocket/mash"
	"go-rocket/metadata"
	"go-rocket/service"

	"log"
)

func main() {
	mash := mash.NewMash()
	mash.BuliderRouter(
		service.BuilderBalance(service.None),
		service.BuildRegMessage(&pb.HelloRequest{}, &pb.HelloReply{}),
		service.BuilderRegCenter(service.NewLocalCenter(map[string]int{
			"127.0.0.1:50051": 1,
		}, []*metadata.URI{
			{
				PackageName:     "proto",
				ServiceName:     "Greeter",
				Method:          "SayHello",
				RequestMessage:  "HelloRequest",
				ResponseMessage: "HelloReply",
			}})))
	err := mash.ListenWithPort(":9000")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
