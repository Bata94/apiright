package core

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Request struct {
	req *http.Request
}

func newRequest(r *http.Request) *Request {
	return &Request{req: r}
}

func (r *Request) Method() string {
	return r.req.Method
}

func (r *Request) URL() *url.URL {
	return r.req.URL
}

func (r *Request) Path() string {
	return r.req.URL.Path
}

func (r *Request) RequestURI() string {
	return r.req.RequestURI
}

func (r *Request) RemoteAddr() string {
	return r.req.RemoteAddr
}

func (r *Request) Header() http.Header {
	return r.req.Header
}

func (r *Request) Body() io.ReadCloser {
	return r.req.Body
}

func (r *Request) Context() context.Context {
	return r.req.Context()
}

func (r *Request) WithContext(ctx context.Context) *http.Request {
	return r.req.WithContext(ctx)
}

func (r *Request) PathValue(name string) string {
	return r.req.PathValue(name)
}

func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return r.req.FormFile(key)
}

func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.req.Cookie(name)
}

func (r *Request) BasicAuth() (string, string, bool) {
	return r.req.BasicAuth()
}

func (r *Request) Proto() string {
	return r.req.Proto
}

func (r *Request) ProtoMajor() int {
	return r.req.ProtoMajor
}

func (r *Request) ProtoMinor() int {
	return r.req.ProtoMinor
}

func (r *Request) Host() string {
	return r.req.Host
}

func (r *Request) Referer() string {
	return r.req.Referer()
}

func (r *Request) UserAgent() string {
	return r.req.UserAgent()
}

func (r *Request) ParseForm() error {
	return r.req.ParseForm()
}

func (r *Request) ParseMultipartForm(maxMemory int64) error {
	return r.req.ParseMultipartForm(maxMemory)
}

func (r *Request) Form() url.Values {
	return r.req.Form
}

func (r *Request) PostForm() url.Values {
	return r.req.PostForm
}

func (r *Request) MultipartForm() *multipart.Form {
	return r.req.MultipartForm
}

func (r *Request) GetBody() (io.ReadCloser, error) {
	return r.req.GetBody()
}

func (r *Request) Close() bool {
	return r.req.Close
}

func (r *Request) TransferEncoding() []string {
	return r.req.TransferEncoding
}

func (r *Request) Unexported() *http.Request {
	return r.req
}
