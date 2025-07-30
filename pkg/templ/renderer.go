package ar_templ

import (
	"github.com/a-h/templ"
	ar "github.com/bata94/apiright/pkg/core"
)

func SimpleRenderer(path string, comp templ.Component) (string, ar.Handler) {
	handler := func(ctx *ar.Ctx) error {
		buf := templ.GetBuffer()
		defer templ.ReleaseBuffer(buf)

		if err := comp.Render(ctx.Request.Context(), buf); err != nil {
			return err
		}

		ctx.Response.SetData(buf.Bytes())
		ctx.Response.AddHeader("Content-Type", "text/html; charset=utf-8")
		return nil
	}

	return path, handler
}
