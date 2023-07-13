package mash

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-rocket/mash/codec"
	meta "go-rocket/metadata"
	"go-rocket/service"
	"go-rocket/ware"
	"log"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type HttpMash struct {
	port          string   // mash port and ip
	headerfilter  []string //use for keep origin header key data from the http request
	routerservice *service.RouterService
	handler       ware.HandlerUnit
	middlewares   []ware.Middleware
	afterhandler  ware.AfterUnit
}

func NewHttpMash() *HttpMash {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	return &HttpMash{
		middlewares:  make([]ware.Middleware, 0),
		headerfilter: make([]string, 0),
	}
}

func (m *HttpMash) AddHeaderfiler(names ...string) {
	m.headerfilter = append(m.headerfilter, names...)
}

func (m *HttpMash) BuliderRouter(builders ...service.RegBuilder) {
	m.BindRouter(service.BuildService(builders...))
}

func (m *HttpMash) BindRouter(routerservice *service.RouterService) {
	m.routerservice = routerservice
	m.AddMiddlewares(m.routerservice.MatcherUnit().WareBuild())
}

func (m *HttpMash) SetListenPort(port string) {
	m.port = port
}

func (m *HttpMash) AddMiddlewares(middleware ...ware.Middleware) {
	m.middlewares = append(m.middlewares, middleware...)
}

func (m *HttpMash) AddAfterHandle(afterware ware.AfterUnit) {
	m.afterhandler = afterware
}

func (m *HttpMash) Listen() error {
	//check the proto message table
	dic := m.routerservice.GetDic()
	if len(dic) == 0 {
		panic(errors.New("no proto message"))
	}
	m.handler = func(ctx context.Context, data *meta.MetaData) (response any, err error) {
		opt := grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.DefaultGRPCCodecs["application/proto"]), grpc.WaitForReady(false))
		//connection by grpc
		gconn, err := grpc.Dial(data.GetHost(), opt, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		//build the grpc metadata
		//head filter
		md := metadata.MD{}
		for _, v := range m.headerfilter {
			if peek := string(data.Request.Header.Peek(v)); len(peek) > 0 {
				md.Append(v, peek)
			}
		}
		context := metadata.NewOutgoingContext(ctx, md)

		//invoke the server moethod by grpc
		in, out := data.GetProtoMessage(dic)
		err = gconn.Invoke(context, data.Descriptor.GetFullMethod(), in, out)
		return out, err
	}

	for _, v := range m.middlewares {
		m.handler = v(m.handler)
	}

	mainRouter := func(reqctx *fasthttp.RequestCtx) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		data := &meta.MetaData{
			RequestCtx: reqctx,
		}
		err := data.FormatAll()
		if err != nil {
			log.Println(err)
			errmsg := meta.NewError("sysem error")
			errmsg.PrintErrorByHttp(reqctx)
			return
		}
		result, err := m.handler(ctx, data)
		if err == nil && m.afterhandler != nil {
			after := &meta.AfterMetaData{
				MetaData: data,
				Result:   result,
			}
			result, _ = m.afterhandler(after)
		}
		if err != nil {
			log.Println(err)
		}
		b, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
			errmsg := meta.NewError("sysem error")
			errmsg.PrintErrorByHttp(reqctx)
		} else {
			reqctx.SetContentType("application/json")
			reqctx.SetBody(b)
		}
	}
	handler := func(reqctx *fasthttp.RequestCtx) {
		switch string(reqctx.Path()) {
		case "/watcher":
			m.routerservice.Watcher(reqctx)
		default:
			mainRouter(reqctx)
		}
	}
	sever := &fasthttp.Server{
		Handler:               handler,
		NoDefaultServerHeader: true,
	}
	return sever.ListenAndServe(m.port)
}

func (m *HttpMash) ListenWithPort(port string) error {
	m.SetListenPort(port)
	return m.Listen()
}

type MashContainer struct {
	HttpMash *HttpMash
	GrpcMash *GrpcMash
}

func NewMashContainer(httpport, grpcport string, builders ...service.RegBuilder) *MashContainer {
	httpMash := NewHttpMash()
	grpcMash := NewGrpcMash()
	router := service.BuildService(builders...)
	httpMash.BindRouter(router)
	grpcMash.BindRouter(router)
	httpMash.port = httpport
	grpcMash.port = grpcport
	return &MashContainer{
		HttpMash: httpMash,
		GrpcMash: grpcMash,
	}
}

func (builder *MashContainer) Listen() {
	ret := make(chan error)
	go func() {
		err := builder.HttpMash.Listen()
		ret <- err
	}()
	go func() {
		err := builder.GrpcMash.Listen()
		ret <- err
	}()

	for e := range ret {
		log.Fatal("ListenAndServe: ", e)
	}
}
