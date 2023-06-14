package main

import (
	pb "go-rocket/example/proto/hello"
	"go-rocket/mash"
	"go-rocket/metadata"
	"go-rocket/service"
	"reflect"

	"log"
)

func main() {
	//fmt.Println(reflect.TypeOf(&pb.HelloRequest{}).Elem().PkgPath())
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
				RequestMessage:  reflect.TypeOf(pb.HelloRequest{}).String(),
				ResponseMessage: reflect.TypeOf(pb.HelloReply{}).String(),
			}})))
	err := mash.ListenWithPort(":9000")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
