type {{ $.InterfaceName }} interface {
{{range .MethodSet}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end}}
}
func Register{{ $.InterfaceName }}(r gin.IRouter, srv {{ $.InterfaceName }}) {
	s := {{.Name}}{
		server: srv,
		router:     r,
	}
	s.RegisterService()
}

type {{$.Name}} struct{
	server {{ $.InterfaceName }}
	router gin.IRouter
}

{{range .Methods}}
func (s *{{$.Name}}) {{ .HandlerName }} (ctx *gin.Context) {
	var in {{.Request}}
{{if .HasPathParams }}
	if err := ginext.ShouldBindUri(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}
{{end}}
	if err := ginext.ShouldBind(ctx, &in); err != nil {
	    ginext.ErrorResponse(ctx, err)
		return
	}
	md := metadata.New(nil)
	md.Set("url", ctx.Request.URL.String())
	md.Set("path", ctx.Request.URL.Path)
	md.Set("query", ctx.Request.URL.Query().Encode())
	md.Set("raw_query", ctx.Request.URL.RawQuery)
	for k, v := range ctx.Request.Header {
		md.Set(k, v...)
	}
	newCtx := metadata.NewIncomingContext(ctx.Request.Context(), md)
	out, err := s.server.({{ $.InterfaceName }}).{{.Name}}(newCtx, &in)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}
{{end}}

func (s *{{$.Name}}) RegisterService() {
{{range .Methods}}
		s.router.Handle("{{.Method}}", "{{.Path}}", s.{{ .HandlerName }})
{{end}}
}