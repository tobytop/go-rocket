package service

import (
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
	PackageName string
	ServiceName string
	Method      string
	Host        string
	InMessage   proto.Message
	OutMessage  proto.Message
}

type Router struct {
	Descriptors map[string]*metadata.Descriptor
	Hosts       map[string]int
}

type LocalCenter struct {
	*Router
}

func NewLocalCenter(hosts map[string]int, path []*RouterInfo) *LocalCenter {
	descriptors := make(map[string]*metadata.Descriptor)
	for _, v := range path {
		p := &metadata.Descriptor{
			URI: &metadata.URI{
				Host:        v.Host,
				Method:      v.Method,
				PackageName: v.PackageName,
				ServiceName: v.ServiceName,
			},
		}
		p.RequestMessage = reflect.TypeOf(v.InMessage).Elem().String()
		p.ResponseMessage = reflect.TypeOf(v.OutMessage).Elem().String()
		key := fmt.Sprintf("/%v.%v/%v", v.PackageName, v.ServiceName, v.Method)
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

func NewLocalCenterNoHost(path []*RouterInfo) *LocalCenter {
	return NewLocalCenter(nil, path)
}

func (l *LocalCenter) LoadDic() (map[string]int, map[string]*metadata.Descriptor) {
	return l.Hosts, l.Descriptors
}

func (l *LocalCenter) Watcher(sender *RegContext) {
}
