package mash

import (
	"go-rocket/ware"

	"google.golang.org/grpc"
)

type GrpcMash struct {
	middlewares []ware.GrpcWare
	handler     grpc.StreamHandler
}

func NewGrpcMash() *GrpcMash {
	return &GrpcMash{
		middlewares: make([]ware.GrpcWare, 0),
		handler:     mainhandler(),
	}
}

func (m *GrpcMash) BulidMash() *grpc.Server {
	for _, v := range m.middlewares {
		m.handler = v(m.handler)
	}
	return grpc.NewServer(grpc.UnknownServiceHandler(m.handler))
}

func (m *GrpcMash) AddMiddlewares(middleware ...ware.GrpcWare) {
	m.middlewares = append(m.middlewares, middleware...)
}

func mainhandler() grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		grpc.MethodFromServerStream(stream)
		return nil
	}
}
