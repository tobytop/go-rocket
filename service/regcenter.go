package service

import (
	"fmt"
	"go-rocket/metadata"
	"net/http"
	"reflect"
	"strings"
)

const (
	Local = iota
	Etcd
	Consul
)

type RegCenter interface {
	LoadDic() (map[string]int, map[string]*metadata.URI)
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
	Paths map[string]*metadata.URI
	Hosts map[string]int
}

type LocalCenter struct {
	*Router
}

func NewLocalCenter(hosts map[string]int, path []*RouterInfo) *LocalCenter {
	paths := make(map[string]*metadata.URI)
	for _, v := range path {
		p := new(metadata.URI)
		p.Method = v.Method
		p.PackageName = v.PackageName
		p.ServiceName = v.ServiceName
		p.RequestMessage = getTypeName(reflect.TypeOf(v.InMessage))
		p.ResponseMessage = getTypeName(reflect.TypeOf(v.OutMessage))
		key := fmt.Sprintf("/%v.%v/%v", strings.ToLower(v.PackageName), strings.ToLower(v.ServiceName), strings.ToLower(v.Method))
		paths[key] = p
	}

	return &LocalCenter{
		Router: &Router{
			Paths: paths,
			Hosts: hosts,
		},
	}
}

func NewLocalCenterNoHost(path []*RouterInfo) *LocalCenter {
	return NewLocalCenter(nil, path)
}

func getTypeName(objType reflect.Type) string {
	if objType.Kind() == reflect.Ptr {
		return objType.Elem().String()
	} else {
		return objType.String()
	}
}

func (l *LocalCenter) LoadDic() (map[string]int, map[string]*metadata.URI) {
	return l.Hosts, l.Paths
}

func (l *LocalCenter) Watcher(sender *RegContext) {
}
