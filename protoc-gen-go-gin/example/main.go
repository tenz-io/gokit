package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	v1 "example/api/product/app/v1"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext"
	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/ginext/metadata"
	"github.com/tenz-io/gokit/logger"
)

var (
	_ v1.BlogServiceHTTPServer = (*service)(nil)
)

type service struct {
}

func (s *service) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginResp, error) {
	// mock login
	if req.GetUsername() != "admin" || req.GetPassword() != "admin" {
		return nil, errcode.Unauthorized(http.StatusUnauthorized, "invalid username or password")
	}

	expiredAt := time.Now().Add(15 * time.Minute)
	token, err := ginext.GenerateToken(123, ginext.RoleAdmin, expiredAt)
	if err != nil {
		return nil, errcode.InternalServer(http.StatusInternalServerError, "failed to generate token")
	}

	return &v1.LoginResp{
		Token: token,
	}, nil
}

func (s *service) CreateArticle(ctx context.Context, req *v1.CreateArticleReq) (*v1.CreateArticleResp, error) {
	log.Printf("CreateArticle: %+v\n", req)

	var (
		meta, existing = metadata.FromContext(ctx)
	)

	log.Printf("existing: %t, username: %s, role: %s\n", existing, meta.Username, meta.Role)

	return &v1.CreateArticleResp{
		ArticleId: 123,
		Title:     req.GetTitle(),
		Content:   req.GetContent(),
	}, nil
}

func (s *service) GetArticles(ctx context.Context, req *v1.GetArticlesReq) (*v1.GetArticlesResp, error) {
	log.Printf("GetArticles: %+v\n", req)

	return &v1.GetArticlesResp{
		Total: int64(req.GetPageSize()),
		Articles: []*v1.Article{
			{
				ArticleId: 123,
				AuthorId:  req.GetAuthorId(),
				Title:     "Hello World",
				Content:   "World World, Hello World",
			},
		},
	}, nil
}

func (s *service) GetImage(ctx context.Context, req *v1.GetImageReq) (*v1.GetImageResp, error) {
	log.Printf("GetImage: %+v\n", req)
	fileContent := []byte("image content")
	return &v1.GetImageResp{
		File: fileContent,
	}, nil
}

func (s *service) UploadImage(ctx context.Context, req *v1.UploadImageReq) (*v1.UploadImageResp, error) {
	log.Printf("UploadImage, key=%s, region: %s, file size: %d\n",
		req.GetKey(), req.GetRegion(), len(req.GetImage()))

	if !strings.HasPrefix(req.GetAuthorization(), "Bearer ") {
		return &v1.UploadImageResp{}, errcode.Forbidden(http.StatusForbidden, "no permission")
	}

	return &v1.UploadImageResp{
		Key: req.GetKey(),
	}, nil
}

func init() {
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
}

func main() {
	e := gin.Default()
	v1.RegisterBlogServiceHTTPServer(e, &service{})
	if err := e.Run(":8888"); err != nil {
		log.Fatal(err)
	}
}
