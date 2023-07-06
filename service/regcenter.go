package service

import (
	"errors"
	"fmt"
	"go-rocket/metadata"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
)

const (
	Local = iota
	Etcd
	Consul
)

type RegCenter interface {
	LoadDic() (map[string]int, map[string]*metadata.Descriptor)
	Watcher(*RegContext)
}

type RegContext struct {
	*Router
	AfterLoad Balance
	Req       *http.Request
	Writer    http.ResponseWriter
}

type RouterInfo struct {
	Path       string
	Host       string
	InMessage  proto.Message
	OutMessage proto.Message
}

type Router struct {
	Descriptors map[string]*metadata.Descriptor
	Hosts       map[string]int
}

type LocalCenter struct {
	*Router
}

func NewLocalCenter(hosts map[string]int, info []*RouterInfo) RegCenter {
	descriptors := make(map[string]*metadata.Descriptor)
	for _, v := range info {
		routerinfo := strings.Split(v.Path, "/")
		if len(routerinfo) != 3 {
			panic(errors.New("wrong url pattern"))
		}
		p := &metadata.Descriptor{
			URI: &metadata.URI{
				Host:        v.Host,
				Method:      routerinfo[2],
				PackageName: routerinfo[0],
				ServiceName: routerinfo[1],
			},
		}
		p.RequestMessage = reflect.TypeOf(v.InMessage).Elem().String()
		p.ResponseMessage = reflect.TypeOf(v.OutMessage).Elem().String()
		key := fmt.Sprintf("/%v.%v/%v", p.PackageName, p.ServiceName, p.Method)
		key = strings.ToLower(key)
		descriptors[key] = p
	}

	return &LocalCenter{
		Router: &Router{
			Descriptors: descriptors,
			Hosts:       hosts,
		},
	}
}

func NewLocalCenterNoHost(path []*RouterInfo) RegCenter {
	return NewLocalCenter(nil, path)
}

func (l *LocalCenter) LoadDic() (map[string]int, map[string]*metadata.Descriptor) {
	return l.Hosts, l.Descriptors
}

func (l *LocalCenter) Watcher(sender *RegContext) {
}
