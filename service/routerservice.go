package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"go-rocket/metadata"
	"go-rocket/ware"
)

type servhost struct {
	*metadata.URI
	addr string
}

type RegBuilder func(*RouterService)

func BuilderBalance(balancetype int) RegBuilder {
	return func(rs *RouterService) {
		var balance Balance
		switch balancetype {
		case RoundRobin:
			balance = &roundRobinBalance{}
		case WeightRobin:
			balance = &weightRoundRobinBalance{}
		default:
			balance = nil
		}
		rs.balance = balance
	}
}
func BuilderRegCenter(reg RegCenter) RegBuilder {
	return func(rs *RouterService) {
		hosts, path := reg.LoadDic()
		rs.Router = &Router{
			Paths: path,
			Hosts: hosts,
		}
		rs.reg = reg
	}
}

type RouterService struct {
	*Router
	balance Balance
	reg     RegCenter
}

func BuildService(builders ...RegBuilder) *RouterService {
	sevice := &RouterService{}
	for _, builder := range builders {
		builder(sevice)
	}

	if sevice.balance != nil {
		for k, v := range sevice.Hosts {
			sevice.balance.Add(k, v)
		}
	}

	return sevice
}

func (s *RouterService) BuildUnit() ware.HandlerUnit {
	return func(ctx context.Context, data *metadata.MetaData) (response any, err error) {
		var uri *metadata.URI
		for _, v := range s.Paths {
			if strings.EqualFold(data.Uri.PackageName, v.PackageName) && strings.EqualFold(data.Uri.Method, v.Method) && strings.EqualFold(data.Uri.ServiceName, v.ServiceName) {
				uri = v
				break
			}
		}
		if uri == nil {
			return nil, errors.New("no router here")
		}
		host := &servhost{
			URI: uri,
		}

		if len(s.Hosts) > 1 && s.balance != nil {
			host.addr = s.balance.next()
		} else {
			for k, _ := range s.Hosts {
				host.addr = k
				break
			}
		}
		data.SetServerHost(host.addr, host.URI)
		return data, nil
	}
}

func (s *RouterService) Watcher(w http.ResponseWriter, r *http.Request) {
	s.reg.Watcher(&RegContext{
		Router:    s.Router,
		AfterLoad: s.balance,
		Writer:    w,
		Req:       r,
	})
}
