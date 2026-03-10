package core

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Request struct {
	Req *http.Request
}

func newRequest(r *http.Request) *Request {
	return &Request{Req: r}
}

func (r *Request) Method() string {
	return r.Req.Method
}

func (r *Request) URL() *url.URL {
	return r.Req.URL
}

func (r *Request) Path() string {
	return r.Req.URL.Path
}

func (r *Request) RequestURI() string {
	return r.Req.RequestURI
}

func (r *Request) RemoteAddr() string {
	return r.Req.RemoteAddr
}

func (r *Request) Header() http.Header {
	return r.Req.Header
}

func (r *Request) Body() io.ReadCloser {
	return r.Req.Body
}

func (r *Request) Context() context.Context {
	return r.Req.Context()
}

func (r *Request) WithContext(ctx context.Context) *http.Request {
	return r.Req.WithContext(ctx)
}

func (r *Request) PathValue(name string) string {
	return r.Req.PathValue(name)
}

func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return r.Req.FormFile(key)
}

func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.Req.Cookie(name)
}

func (r *Request) BasicAuth() (string, string, bool) {
	return r.Req.BasicAuth()
}

func (r *Request) Proto() string {
	return r.Req.Proto
}

func (r *Request) ProtoMajor() int {
	return r.Req.ProtoMajor
}

func (r *Request) ProtoMinor() int {
	return r.Req.ProtoMinor
}

func (r *Request) Host() string {
	return r.Req.Host
}

func (r *Request) Referer() string {
	return r.Req.Referer()
}

func (r *Request) UserAgent() string {
	return r.Req.UserAgent()
}

func (r *Request) ParseForm() error {
	return r.Req.ParseForm()
}

func (r *Request) ParseMultipartForm(maxMemory int64) error {
	return r.Req.ParseMultipartForm(maxMemory)
}

func (r *Request) Form() url.Values {
	return r.Req.Form
}

func (r *Request) PostForm() url.Values {
	return r.Req.PostForm
}

func (r *Request) MultipartForm() *multipart.Form {
	return r.Req.MultipartForm
}

func (r *Request) GetBody() (io.ReadCloser, error) {
	return r.Req.GetBody()
}

func (r *Request) Close() bool {
	return r.Req.Close
}

func (r *Request) TransferEncoding() []string {
	return r.Req.TransferEncoding
}

func (r *Request) Raw() *http.Request {
	return r.Req
}
