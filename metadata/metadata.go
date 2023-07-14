package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type MetaData struct {
	*fasthttp.RequestCtx
	Descriptor *Descriptor
	Payload    map[string]any
	Header     *metadata.MD
	serverhost string
}

type AfterMetaData struct {
	*MetaData
	Result any
}

type ErrorMeta struct {
	Error string `json:"error"`
}

func NewError(message string) *ErrorMeta {
	return &ErrorMeta{
		Error: message,
	}
}

func (err *ErrorMeta) PrintErrorByHttp(ctx *fasthttp.RequestCtx) {
	b, _ := json.Marshal(err)
	ctx.SetContentType("application/json")
	ctx.SetBody(b)
}

func PrintDefaultError(ctx *fasthttp.RequestCtx, err any) {
	log.Println(err)
	errmsg := NewError("sysem error")
	errmsg.PrintErrorByHttp(ctx)
}

type URI struct {
	PackageName string
	ServiceName string
	Method      string
	Host        string
}

type Descriptor struct {
	*URI
	RequestMessage  string
	ResponseMessage string
}

func (d *Descriptor) convertToMessage(dic map[string]proto.Message) (proto.Message, proto.Message) {
	reqIn := dic[d.RequestMessage]
	resOut := dic[d.ResponseMessage]
	req := reflect.New(reflect.TypeOf(reqIn).Elem()).Interface()
	res := reflect.New(reflect.TypeOf(resOut).Elem()).Interface()

	in := req.(proto.Message)
	out := res.(proto.Message)
	return in, out
}

func (d *Descriptor) GetFullMethod() string {
	return fmt.Sprintf("/%v.%v/%v", d.PackageName, d.ServiceName, d.Method)
}

func (m *MetaData) SetServerHost(host string) {
	m.serverhost = host
}

func (m *MetaData) GetHost() string {
	return m.serverhost
}
func (m *MetaData) FormatAll() error {
	err := m.formatUri()
	if err == nil {
		m.FormatPayload()
	}
	return err
}

func (m *MetaData) formatUri() error {
	path := string(m.Path())
	st := strings.Split(path, "/")
	if len(st) != 4 {
		return errors.New(path + " is wrong url")
	}
	m.Descriptor = &Descriptor{
		URI: &URI{
			PackageName: st[1],
			ServiceName: st[2],
			Method:      st[3],
		},
	}
	return nil
}

func (m *MetaData) FormatPayload() {
	m.Payload = make(map[string]any)
	m.QueryArgs().VisitAll(m.visitallquery)
	m.PostArgs().VisitAll(m.visitallquery)
	if len(m.Payload) == 0 {
		if err := json.Unmarshal(m.PostBody(), &m.Payload); err != nil {
			log.Println(err)
			log.Println(string(m.PostBody()))
		}
	}
}

func (m *MetaData) visitallquery(key, value []byte) {
	mainkey := string(key)
	data := string(value)
	m.Payload[mainkey] = data
}

func (m *MetaData) GetProtoMessage(dic map[string]proto.Message) (proto.Message, proto.Message) {
	if len(dic) == 0 {
		return nil, nil
	}
	reqIn, resOut := m.Descriptor.convertToMessage(dic)
	err := mapstructure.Decode(m.Payload, reqIn)

	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	return reqIn, resOut
}
