package mash

import (
	"context"
	"encoding/json"
	"fmt"
	"go-rocket/mash/codec"
	meta "go-rocket/metadata"
	"go-rocket/service"
	"go-rocket/ware"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Mash struct {
	httpPort      string
	routerservice *service.RouterService
	handler       ware.HandlerUnit
	middlewares   []ware.Middleware
	afterhandler  ware.AfterUnit
}

func NewMash() *Mash {
	return &Mash{
		middlewares: make([]ware.Middleware, 0),
	}
}

func (m *Mash) BuliderRouter(builders ...service.RegBuilder) {
	m.routerservice = service.BuildService(builders...)
	m.AddMiddlewares(m.routerservice.BuildUnit().WareBuild())
}

func (m *Mash) SetListenPort(port string) {
	m.httpPort = port
}

func (m *Mash) AddMiddlewares(middleware ...ware.Middleware) {
	m.middlewares = append(m.middlewares, middleware...)
}

func (m *Mash) AddAfterHandle(afterware ware.AfterUnit) {
	m.afterhandler = afterware
}

func (m *Mash) Listen() error {
	m.handler = func(ctx context.Context, data *meta.MetaData) (response any, err error) {

		opt := grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.DefaultGRPCCodecs["application/json"]), grpc.WaitForReady(false))
		//connection by grpc
		gconn, err := grpc.Dial(data.GetHost(), opt, grpc.WithInsecure())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		//build the grpc metadata
		md := metadata.MD{}
		context := metadata.NewOutgoingContext(ctx, md)
		callopt := grpc.Header(data.Header)

		//invoke the server moethod by grpc
		var out any
		err = gconn.Invoke(context, data.Uri.GetFullMethod(), data.Params, &out, callopt)
		return out, err
	}

	for _, v := range m.middlewares {
		m.handler = v(m.handler)
	}

	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		data := &meta.MetaData{
			Req: r,
		}
		err := data.FormatAll()
		if err != nil {
			errmsg := meta.NewError(err)
			errmsg.PrintErrorByHttp(w)
			return
		}

		result, err := m.handler(ctx, data)
		if err != nil {
			errmsg := meta.NewError(err)
			errmsg.PrintErrorByHttp(w)
			return
		}

		if m.afterhandler != nil {
			after := &meta.AfterMetaData{
				MetaData: data,
				Result:   result,
			}
			result, _ = m.afterhandler(after)
		}

		b, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	mux.HandleFunc("/watcher", m.routerservice.Watcher)
	return http.ListenAndServe(m.httpPort, mux)
}

func (g *Mash) ListenWithPort(port string) error {
	g.SetListenPort(port)
	return g.Listen()
}
