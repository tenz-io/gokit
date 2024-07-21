package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"example/api/example/v1"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext"
	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/logger"
)

var (
	_ v1.ApiServerHTTPServer = (*service)(nil)
)

type service struct {
}

func (s *service) UploadImage(ctx context.Context, request *v1.UploadImageRequest) (*v1.UploadImageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetImage(ctx context.Context, request *v1.GetImageRequest) (*v1.GetImageResponse, error) {
	log.Printf("GetImage: %+v\n", request)
	fileContent := []byte("image content")
	return &v1.GetImageResponse{
		File: fileContent,
	}, nil
}

func (s *service) Hello(ctx context.Context, request *v1.HelloRequest) (*v1.HelloResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// mock login
	if req.GetUsername() != "admin" || req.GetPassword() != "admin" {
		return nil, errcode.Unauthorized(http.StatusUnauthorized, "invalid username or password")
	}

	expiredAt := time.Now().Add(15 * time.Minute)
	accessToken, err := ginext.GenerateToken(123, ginext.RoleAdmin, ginext.TokenTypeAccess, expiredAt)
	if err != nil {
		return nil, errcode.InternalServer(http.StatusInternalServerError, "failed to generate token")
	}
	refreshToken, err := ginext.GenerateToken(123, ginext.RoleAdmin, ginext.TokenTypeRefresh, time.Now().Add(365*24*time.Hour))
	if err != nil {
		return nil, errcode.InternalServer(http.StatusInternalServerError, "failed to generate token")
	}

	return &v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Query(ctx context.Context, request *v1.QueryRequest) (*v1.QueryResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) UpdateProgress(ctx context.Context, request *v1.UpdateProgressRequest) (*v1.UpdateProgressResponse, error) {
	//TODO implement me
	panic("implement me")
}

func init() {
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
}

func main() {
	e := gin.Default()
	v1.RegisterApiServerHTTPServer(e, &service{})
	if err := e.Run(":8888"); err != nil {
		log.Fatal(err)
	}
}
