package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/functions/metadata"
	v1 "example/api/product/app/v1"

	"github.com/gin-gonic/gin"
)

type service struct {
}

func (s *service) UploadImage(ctx context.Context, req *v1.UploadImageReq) (*v1.UploadImageResp, error) {
	if len(req.GetFile()) == 0 {
		return nil, fmt.Errorf("file is required")
	}
	if req.GetFilename() == "" {
		return nil, fmt.Errorf("filename is required")
	}

	return &v1.UploadImageResp{
		Key: fmt.Sprintf("%d-%s", len(req.GetFile()), req.GetFilename()),
	}, nil
}

func (s *service) CreateArticle(ctx context.Context, article *v1.Article) (*v1.Article, error) {
	var (
		meta, err = metadata.FromContext(ctx)
	)
	if err != nil {
		return nil, fmt.Errorf("get metadata failed: %w", err)
	}

	log.Printf("metadata: %+v", meta)

	if article.AuthorId < 1 {
		return nil, fmt.Errorf("author id must > 0")
	}
	return article, nil
}

func (s *service) GetArticles(ctx context.Context, req *v1.GetArticlesReq) (*v1.GetArticlesResp, error) {
	if req.AuthorId < 0 {
		return nil, fmt.Errorf("author id must >= 0")
	}
	return &v1.GetArticlesResp{
		Total: 1,
		Articles: []*v1.Article{
			{
				Title:    "test article: " + req.Title,
				Content:  "test",
				AuthorId: 1,
			},
		},
	}, nil
}

func main() {
	e := gin.Default()
	v1.RegisterBlogServiceHTTPServer(e, &service{})
	if err := e.Run(":8888"); err != nil {
		log.Fatal(err)
	}
}
