package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type MetaData struct {
	Req        *http.Request
	Descriptor *Descriptor
	Codec      string
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

func NewError(e error) *ErrorMeta {
	return &ErrorMeta{
		Error: e.Error(),
	}
}

func (err *ErrorMeta) PrintErrorByHttp(writer http.ResponseWriter) {
	b, _ := json.Marshal(err)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(b)
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

func (u *URI) GetFullMethod() string {
	return fmt.Sprintf("/%v.%v/%v", u.PackageName, u.ServiceName, u.Method)
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
		m.FormatHeader()
	}
	return err
}

func (m *MetaData) formatUri() error {
	st := strings.Split(m.Req.URL.Path, "/")
	if len(st) != 4 {
		return errors.New("url is wrong")
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
	m.Req.ParseForm()
	payload := make(map[string]any)
	for key, v := range m.Req.Form {
		var data map[string]any
		err := json.Unmarshal([]byte(key), &data)
		if err == nil {
			for kk, vv := range data {
				payload[kk] = vv
			}
		} else {
			if len(v) > 0 {
				payload[key] = v[0]
			} else {
				payload[key] = ""
			}
		}
	}
	m.Payload = payload
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

func (m *MetaData) FormatHeader() {
	if contenttype, ok := m.Req.Header["Request-Content-Type"]; ok && len(contenttype) > 0 {
		m.Codec = contenttype[0]
	} else {
		m.Codec = "application/proto"
	}
	m.Header = (*metadata.MD)(&m.Req.Header)
}
