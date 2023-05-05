package main

import (
	"go-rocket/mash"
	"go-rocket/metadata"
	"go-rocket/service"
)

func main() {
	mash := mash.NewMash(nil)
	mash.BuliderRouter(service.BuilderBalance(service.None), service.BuilderRegCenter(service.NewLocalCenter(map[string]int{
		":8080": 1,
	}, []*metadata.URI{
		{
			PackageName: "test",
			ServiceName: "test",
			Version:     "v1",
			Method:      "helloword",
		}})))
	mash.ListenWithPort("9000")
}
