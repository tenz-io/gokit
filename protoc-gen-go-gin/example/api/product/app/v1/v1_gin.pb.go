// Code generated by github.com/tenz-io/gokit/protoc-gen-go-gin. DO NOT EDIT.

package v1

import (
	context "context"
	gin "github.com/gin-gonic/gin"
	ginext "github.com/tenz-io/gokit/ginext"
	metadata "github.com/tenz-io/gokit/ginext/metadata"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the github.com/tenz-io/gokit/protoc-gen-go-gin package it is being compiled against.
// context.metadata.
// gin.ginext.

type BlogServiceHTTPServer interface {
	CreateArticle(context.Context, *CreateArticleReq) (*CreateArticleResp, error)

	GetArticles(context.Context, *GetArticlesReq) (*GetArticlesResp, error)

	GetImage(context.Context, *GetImageReq) (*GetImageResp, error)

	Login(context.Context, *LoginReq) (*LoginResp, error)

	UploadImage(context.Context, *UploadImageReq) (*UploadImageResp, error)
}

func RegisterBlogServiceHTTPServer(r gin.IRouter, srv BlogServiceHTTPServer) {
	s := BlogService{
		server: srv,
		router: r,
	}
	s.RegisterService()
}

type BlogService struct {
	server BlogServiceHTTPServer
	router gin.IRouter
}

func (s *BlogService) Login_0(ctx *gin.Context) {
	var in LoginReq
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
		return s.server.(BlogServiceHTTPServer).Login(ctx, req.(*LoginReq))
	}

	md := metadata.New(ctx, "BlogServiceHTTPServer.Login")
	newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
	out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}

func (s *BlogService) GetArticles_0(ctx *gin.Context) {
	var in GetArticlesReq
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
		return s.server.(BlogServiceHTTPServer).GetArticles(ctx, req.(*GetArticlesReq))
	}

	md := metadata.New(ctx, "BlogServiceHTTPServer.GetArticles")
	newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
	out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}

func (s *BlogService) CreateArticle_0(ctx *gin.Context) {
	var in CreateArticleReq
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
		return s.server.(BlogServiceHTTPServer).CreateArticle(ctx, req.(*CreateArticleReq))
	}

	md := metadata.New(ctx, "BlogServiceHTTPServer.CreateArticle")
	newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
	out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}

func (s *BlogService) UploadImage_0(ctx *gin.Context) {
	var in UploadImageReq
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
		return s.server.(BlogServiceHTTPServer).UploadImage(ctx, req.(*UploadImageReq))
	}

	md := metadata.New(ctx, "BlogServiceHTTPServer.UploadImage")
	newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
	out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}

func (s *BlogService) GetImage_0(ctx *gin.Context) {
	var in GetImageReq
	if err := ginext.BindAndValidate(ctx, &in); err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	var handler ginext.RpcHandler = func(ctx context.Context, req any) (resp any, err error) {
		return s.server.(BlogServiceHTTPServer).GetImage(ctx, req.(*GetImageReq))
	}

	md := metadata.New(ctx, "BlogServiceHTTPServer.GetImage")
	newCtx := metadata.WithMetadata(ctx.Request.Context(), md)
	out, err := ginext.AllRpcInterceptor.Intercept(newCtx, &in, handler)
	if err != nil {
		ginext.ErrorResponse(ctx, err)
		return
	}

	ginext.Response(ctx, out)
}

func (s *BlogService) RegisterService() {

	s.router.Handle("POST", "/login", ginext.Authenticate(1), s.Login_0)

	s.router.Handle("GET", "/v1/author/:author_id/articles", ginext.Authenticate(2), s.GetArticles_0)

	s.router.Handle("POST", "/v1/author/:author_id/articles", ginext.Authenticate(4), s.CreateArticle_0)

	s.router.Handle("POST", "/v1/images/:key", ginext.Authenticate(4), s.UploadImage_0)

	s.router.Handle("GET", "/v1/images/:key", ginext.Authenticate(2), s.GetImage_0)

}
