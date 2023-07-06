package mash

import (
	"context"
	"errors"
	"go-rocket/mash/codec"
	meta "go-rocket/metadata"
	"go-rocket/service"
	"go-rocket/ware"
	"io"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	clientStreamDescForProxying = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

type GrpcMash struct {
	port          string
	routerservice *service.RouterService
	handler       ware.HandlerUnit
	middlewares   []ware.Middleware
}

func NewGrpcMash() *GrpcMash {
	return &GrpcMash{
		middlewares: make([]ware.Middleware, 0),
	}
}

func (m *GrpcMash) BuliderRouter(builders ...service.RegBuilder) {
	m.BindRouter(service.BuildService(builders...))
}

func (m *GrpcMash) BindRouter(routerservice *service.RouterService) {
	m.routerservice = routerservice
}

func (m *GrpcMash) Listen() error {
	m.handler = m.routerservice.MatcherUnit()
	for _, v := range m.middlewares {
		m.handler = v(m.handler)
	}
	server := grpc.NewServer(grpc.UnknownServiceHandler(m.transhandler()))
	lis, err := net.Listen("tcp", m.port)
	if err != nil {
		return err
	}
	return server.Serve(lis)
}

func (m *GrpcMash) AddMiddlewares(middleware ...ware.Middleware) {
	m.middlewares = append(m.middlewares, middleware...)
}

func (m *GrpcMash) transhandler() grpc.StreamHandler {
	return func(srv interface{}, serverStream grpc.ServerStream) error {
		path, ok := grpc.MethodFromServerStream(serverStream)
		if !ok {
			return errors.New("path is wrong")
		}
		incomingCtx := serverStream.Context()
		clientCtx, clientCancel := context.WithCancel(incomingCtx)
		defer clientCancel()

		header, _ := metadata.FromIncomingContext(clientCtx)
		data := buildMeta(path)
		if data == nil {
			return errors.New(path)
		}
		data.Header = &header
		result, err := m.handler(clientCtx, data)
		if err != nil {
			return err
		}
		handlerdata, ok := result.(*meta.MetaData)
		if !ok {
			return errors.New("the metadata is wrong")
		}
		newCtx := metadata.NewOutgoingContext(clientCtx, handlerdata.Header.Copy())

		//connection by grpc
		opt := grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.DefaultGRPCCodecs["application/proto"]), grpc.WaitForReady(false))
		gconn, err := grpc.Dial(handlerdata.GetHost(), opt, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err)
			return err
		}
		clientStream, err := grpc.NewClientStream(newCtx, clientStreamDescForProxying, gconn, path)
		if err != nil {
			log.Println(err)
			return err
		}
		s2cErrChan := m.forwardServerToClient(serverStream, clientStream)
		c2sErrChan := m.forwardClientToServer(clientStream, serverStream)
		for i := 0; i < 2; i++ {
			select {
			case s2cErr := <-s2cErrChan:
				if s2cErr == io.EOF {
					// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
					// the clientStream>serverStream may continue pumping though.
					clientStream.CloseSend()
				} else {
					// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
					// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
					// exit with an error to the stack
					clientCancel()
					return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
				}
			case c2sErr := <-c2sErrChan:
				// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
				// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
				// will be nil.
				serverStream.SetTrailer(clientStream.Trailer())
				// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
				if c2sErr != io.EOF {
					return c2sErr
				}
				return nil
			}
		}
		return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
	}
}

func (m *GrpcMash) forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for {
			if err := src.RecvMsg(f); err != nil {
				ret <- err
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func (m *GrpcMash) forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func buildMeta(path string) *meta.MetaData {
	str := strings.Split(path[1:], "/")
	if len(str) != 2 {
		return nil
	}
	index := strings.LastIndex(str[0], ".")
	packagename := str[0][:index]
	servername := str[0][index+1:]
	return &meta.MetaData{
		Descriptor: &meta.Descriptor{
			URI: &meta.URI{
				Method:      str[1],
				PackageName: packagename,
				ServiceName: servername,
			},
		},
	}
}
