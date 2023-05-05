package service

import (
	"go-rocket/metadata"
	"net/http"
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

type Router struct {
	Paths []*metadata.URI
	Hosts map[string]int
}

type LocalCenter struct {
	*Router
}

func NewLocalCenter(hosts map[string]int, path []*metadata.URI) *LocalCenter {
	return &LocalCenter{
		Router: &Router{
			Paths: path,
			Hosts: hosts,
		},
	}
}

func (l *LocalCenter) LoadDic() (map[string]int, []*metadata.URI) {
	return l.Hosts, l.Paths
}

func (l *LocalCenter) Watcher(sender *RegContext) {
}
