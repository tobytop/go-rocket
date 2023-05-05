package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

type UrlFormat func(*http.Request) (*URI, error)
type MetaData struct {
	Req        *http.Request
	Uri        *URI
	Params     map[string]any
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
	Version     string
	Method      string
	fullMethod  string
}

func (u *URI) formatUrl() {
	u.fullMethod = fmt.Sprintf("/%v/%v", u.ServiceName, u.Method)
}
func (u *URI) GetFullMethod() string {
	return u.fullMethod
}

func (m *MetaData) SetServerHost(host string) {
	m.serverhost = host
}

func (m *MetaData) GetHost() string {
	return m.serverhost
}
func (m *MetaData) FormatAll(format UrlFormat) error {
	var err error
	if format != nil {
		uri, err := format(m.Req)
		if err != nil {
			m.Uri = uri
		}
	} else {
		err = m.formatUri()
	}
	if err == nil {
		m.FormatParams()
		m.FormatHeader()
	}
	return err
}

func (m *MetaData) formatUri() error {
	st := strings.Split(m.Req.URL.Path, "/")
	if len(st) < 5 {
		return errors.New("url is wrong")
	}
	m.Uri = &URI{
		PackageName: st[1],
		ServiceName: st[2],
		Version:     st[3],
		Method:      st[4],
	}
	m.Uri.formatUrl()
	return nil
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
