package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	v1 "example/api/product/app/v1"

	"github.com/gin-gonic/gin"
)

var (
	_ v1.BlogServiceHTTPServer = (*service)(nil)
)

type service struct {
}

func (s service) CreateArticle(ctx context.Context, req *v1.CreateArticleReq) (*v1.CreateArticleResp, error) {
	log.Printf("CreateArticle: %+v\n", req)

	if !strings.HasPrefix(req.GetAuthorization(), "Bearer ") {
		return &v1.CreateArticleResp{}, fmt.Errorf("no permission")
	}

	return &v1.CreateArticleResp{
		ArticleId: 123,
		Title:     req.GetTitle(),
		Content:   req.GetContent(),
	}, nil
}

func (s service) GetArticles(ctx context.Context, req *v1.GetArticlesReq) (*v1.GetArticlesResp, error) {
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

func (s service) GetImage(ctx context.Context, req *v1.GetImageReq) (*v1.GetImageResp, error) {
	log.Printf("GetImage: %+v\n", req)
	fileContent := []byte("image content")
	return &v1.GetImageResp{
		File: fileContent,
	}, nil
}

func (s service) UploadImage(ctx context.Context, req *v1.UploadImageReq) (*v1.UploadImageResp, error) {
	log.Printf("UploadImage, key=%s, file size: %d\n", req.GetKey(), len(req.GetImage()))

	if !strings.HasPrefix(req.GetAuthorization(), "Bearer ") {
		return &v1.UploadImageResp{}, fmt.Errorf("no permission")
	}

	return &v1.UploadImageResp{
		Key: req.GetKey(),
	}, nil
}

func main() {
	e := gin.Default()
	v1.RegisterBlogServiceHTTPServer(e, &service{})
	if err := e.Run(":8888"); err != nil {
		log.Fatal(err)
	}
}
