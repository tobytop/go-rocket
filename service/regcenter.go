package service

import (
	"go-rocket/metadata"
	"net/http"
	"reflect"
)

const (
	Local = iota
	Etcd
	Consul
)

type RegCenter interface {
	LoadDic() (map[string]int, []*metadata.URI)
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
	InMessage   any
	OutMessage  any
}

type Router struct {
	Paths []*metadata.URI
	Hosts map[string]int
}

type LocalCenter struct {
	*Router
}

func NewLocalCenter(hosts map[string]int, path []*RouterInfo) *LocalCenter {
	paths := make([]*metadata.URI, 0)
	for _, v := range path {
		p := new(metadata.URI)
		p.Method = v.Method
		p.PackageName = v.PackageName
		p.ServiceName = v.ServiceName
		p.RequestMessage = getTypeName(reflect.TypeOf(v.InMessage))
		p.ResponseMessage = getTypeName(reflect.TypeOf(v.OutMessage))
		paths = append(paths, p)
	}

	return &LocalCenter{
		Router: &Router{
			Paths: paths,
			Hosts: hosts,
		},
	}
}

func getTypeName(objType reflect.Type) string {
	if objType.Kind() == reflect.Ptr {
		return objType.Elem().String()
	} else {
		return objType.String()
	}
}

func (l *LocalCenter) LoadDic() (map[string]int, []*metadata.URI) {
	return l.Hosts, l.Paths
}

func (l *LocalCenter) Watcher(sender *RegContext) {
}
