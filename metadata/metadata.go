package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
)

type MetaData struct {
	Req        *http.Request
	Uri        *URI
	Params     map[string]interface{}
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

type RegisterData struct {
	*URI
	Message proto.Message
}

type URI struct {
	PackageName string
	ServiceName string
	Version     string
	Method      string
	fullMethod  string
}

func (u *URI) GetFullMethod() string {
	return u.fullMethod
}

func (m *MetaData) SetServerHost(host string, uri *URI) {
	m.serverhost = host
	if m.Uri.Version == "" {
		m.Uri.fullMethod = fmt.Sprintf("/%v.%v/%v", uri.PackageName, uri.ServiceName, uri.Method)
	} else {
		m.Uri.fullMethod = fmt.Sprintf("/%v.%v.%v/%v", uri.PackageName, uri.Version, uri.ServiceName, uri.Method)
	}
}

func (m *MetaData) GetHost() string {
	return m.serverhost
}
func (m *MetaData) FormatAll() error {
	err := m.formatUri()
	if err == nil {
		m.FormatParams()
		m.FormatHeader()
	}
	return err
}

func (m *MetaData) formatUri() error {
	st := strings.Split(m.Req.URL.Path, "/")
	switch len(st) {
	case 4:
		m.Uri = &URI{
			PackageName: st[1],
			ServiceName: st[2],
			Method:      st[3],
		}
		return nil
	case 5:
		m.Uri = &URI{
			PackageName: st[1],
			ServiceName: st[2],
			Version:     st[3],
			Method:      st[4],
		}
		return nil
	default:
		return errors.New("url is wrong")
	}
}

func (m *MetaData) FormatParams() {
	m.Req.ParseForm()
	params := make(map[string]any)
	for key, v := range m.Req.Form {
		var data map[string]any
		err := json.Unmarshal([]byte(key), &data)
		if err == nil {
			for kk, vv := range data {
				params[kk] = vv
			}
		} else {
			if len(v) > 0 {
				params[key] = v[0]
			} else {
				params[key] = ""
			}
		}
	}
	m.Params = params
}

func (m *MetaData) FormatHeader() {
	m.Header = new(metadata.MD)
}
