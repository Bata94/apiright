package ar_templ

import (
	"github.com/a-h/templ"
	ar "github.com/bata94/apiright/pkg/core"
)

func SimpleRenderer(path string, comp templ.Component, opt ...ar.RouteOption) (string, ar.Handler, []ar.RouteOption) {
	handler := func(ctx *ar.Ctx) error {
		buf := templ.GetBuffer()
		defer templ.ReleaseBuffer(buf)

		return comp.Render(ctx.Request.Context(), buf)
	}

	return path, handler, opt
}
