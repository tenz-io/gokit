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
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
	    ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
        return s.server.({{ $.InterfaceName }}).{{.Name}}(ctx, req.(*{{.Request}}))
    }

	md := metadata.New(ctx, "{{ $.InterfaceName }}.{{.Name}}")
    newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
    out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}
{{end}}

func (s *{{$.Name}}) RegisterService() {
{{range .Methods}}
    s.router.Handle("{{.Method}}", "{{.Path}}", ginext.Authenticate({{.Role}}, {{.AuthType}}), s.{{ .HandlerName }})
{{end}}
}