package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"reflect"

	"go-rocket/metadata"
	"go-rocket/ware"

	"google.golang.org/protobuf/proto"
)

type servhost struct {
	*metadata.URI
	addr string
}

type RegBuilder func(*RouterService)

func BuilderBalance(balancetype int) RegBuilder {
	return func(rs *RouterService) {
		log.Println("begin loading balance")
		rs.balance = NewBalance(balancetype)
		log.Println("balance is" + reflect.ValueOf(rs.balance).Kind().String())
	}
}

func BuilderRegCenter(reg RegCenter) RegBuilder {
	return func(rs *RouterService) {
		log.Println("registor router")
		hosts, path := reg.LoadDic()
		rs.Router = &Router{
			Paths: path,
			Hosts: hosts,
		}
		rs.reg = reg
		log.Println("end registor router")
	}
}

func BuildRegMessage(regs ...proto.Message) RegBuilder {
	return func(rs *RouterService) {
		log.Println("registor grpc message")
		regtable := make(map[string]proto.Message)
		for _, reg := range regs {
			typeOfMessage := reflect.TypeOf(reg)
			typeOfMessage = typeOfMessage.Elem()
			regtable[typeOfMessage.String()] = reg
		}
		rs.regtable = regtable
		log.Println("end registor grpc message")
	}
}

type RouterService struct {
	*Router
	regtable map[string]proto.Message
	balance  Balance
	reg      RegCenter
}

func (s *RouterService) GetDic() map[string]proto.Message {
	return s.regtable
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
		uri, ok := s.Paths[data.Uri.GetFullMethod()]
		if !ok {
			return nil, errors.New("no router here")
		}

		data.Uri.Method = uri.Method
		data.Uri.PackageName = uri.PackageName
		data.Uri.ServiceName = uri.ServiceName
		data.Uri.RequestMessage = uri.RequestMessage
		data.Uri.ResponseMessage = uri.ResponseMessage

		var addr string
		if len(s.Hosts) > 1 && s.balance != nil {
			addr = s.balance.next()
		} else {
			for k := range s.Hosts {
				addr = k
				break
			}
		}
		data.SetServerHost(addr)
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
