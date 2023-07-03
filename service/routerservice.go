package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"

	"go-rocket/metadata"
	"go-rocket/ware"

	"google.golang.org/protobuf/proto"
)

type RegBuilder func(*RouterService)

func BuilderBalance(balancetype int) RegBuilder {
	return func(rs *RouterService) {
		log.Println("begin loading balance")
		rs.balance = NewBalance(balancetype)
		log.Println("balance is" + reflect.ValueOf(rs.balance).Kind().String())
	}
}

func BuilderRegCenter(regcenter RegCenter) RegBuilder {
	return func(rs *RouterService) {
		log.Println("registor router")
		hosts, descriptors := regcenter.LoadDic()
		rs.Router = &Router{
			Descriptors: descriptors,
			Hosts:       hosts,
		}
		rs.regcenter = regcenter
		log.Println("end registor router")
	}
}

func BuildRegMessage(protomessages ...proto.Message) RegBuilder {
	return func(rs *RouterService) {
		log.Println("registor grpc message")
		regtable := make(map[string]proto.Message)
		for _, proto := range protomessages {
			typeOfMessage := reflect.TypeOf(proto)
			typeOfMessage = typeOfMessage.Elem()
			regtable[typeOfMessage.String()] = proto
		}
		rs.regtable = regtable
		log.Println("end registor grpc message")
	}
}

type RouterService struct {
	*Router
	regtable  map[string]proto.Message
	balance   Balance
	regcenter RegCenter
}

func (rs *RouterService) GetDic() map[string]proto.Message {
	return rs.regtable
}

func BuildService(builders ...RegBuilder) *RouterService {
	sevice := &RouterService{}
	for _, builder := range builders {
		builder(sevice)
	}

	if sevice.balance != nil && sevice.Hosts != nil {
		for k, v := range sevice.Hosts {
			sevice.balance.Add(k, v)
		}
	}

	return sevice
}

func (rs *RouterService) MatcherUnit() ware.HandlerUnit {
	return func(ctx context.Context, data *metadata.MetaData) (any, error) {
		key := strings.ToLower(data.Descriptor.GetFullMethod())
		log.Println(key)
		descriptor, ok := rs.Descriptors[key]
		if !ok {
			return nil, errors.New("no router here")
		}

		data.Descriptor.Method = descriptor.Method
		data.Descriptor.PackageName = descriptor.PackageName
		data.Descriptor.ServiceName = descriptor.ServiceName
		data.Descriptor.RequestMessage = descriptor.RequestMessage
		data.Descriptor.ResponseMessage = descriptor.ResponseMessage

		var addr string
		if rs.Hosts == nil || len(rs.Hosts) == 0 {
			addr = descriptor.Host
		} else if len(rs.Hosts) > 1 && rs.balance != nil {
			addr = rs.balance.next()
		} else {
			for k := range rs.Hosts {
				addr = k
				break
			}
		}

		data.SetServerHost(addr)
		return data, nil
	}
}

func (rs *RouterService) Watcher(w http.ResponseWriter, r *http.Request) {
	rs.regcenter.Watcher(&RegContext{
		Router:    rs.Router,
		AfterLoad: rs.balance,
		Writer:    w,
		Req:       r,
	})
}
