// Code generated by github.com/tenz-io/gokit/protoc-gen-go-validator. DO NOT EDIT.

package v1

import (
	fmt "fmt"
	genproto "github.com/tenz-io/gokit/genproto"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the github.com/tenz-io/gokit/protoc-gen-go-validator package it is being compiled against.
// fmt.
// genproto.
// strings.

func init() {
	fmt.Sprint("")
	genproto.IsNilOrEmpty(nil)
	strings.TrimSpace("")
}

// message
func (x *LoginRequest) Validate() error {

	if err := x.validateUsername(); err != nil {
		return err
	}

	if err := x.validatePassword(); err != nil {
		return err
	}

	return nil
}

func (x *LoginRequest) validateUsername() error {

	if genproto.IsNilOrEmpty(x.Username) {
		return &genproto.ValidationError{
			Key:     "Username",
			Message: "is required",
		}
	}

	if strings.TrimSpace(x.GetUsername()) == "" {
		return &genproto.ValidationError{
			Key:     "Username",
			Message: "can not be blank",
		}
	}

	if len(x.GetUsername()) < 2 {
		return &genproto.ValidationError{
			Key:     "Username",
			Message: fmt.Sprintf("must be at least %d characters long", 2),
		}
	}

	if len(x.GetUsername()) > 64 {
		return &genproto.ValidationError{
			Key:     "Username",
			Message: fmt.Sprintf("must be at most %d characters long", 64),
		}
	}

	return nil
}

func (x *LoginRequest) validatePassword() error {

	if genproto.IsNilOrEmpty(x.Password) {
		return &genproto.ValidationError{
			Key:     "Password",
			Message: "is required",
		}
	}

	if strings.TrimSpace(x.GetPassword()) == "" {
		return &genproto.ValidationError{
			Key:     "Password",
			Message: "can not be blank",
		}
	}

	if len(x.GetPassword()) < 6 {
		return &genproto.ValidationError{
			Key:     "Password",
			Message: fmt.Sprintf("must be at least %d characters long", 6),
		}
	}

	if len(x.GetPassword()) > 64 {
		return &genproto.ValidationError{
			Key:     "Password",
			Message: fmt.Sprintf("must be at most %d characters long", 64),
		}
	}

	return nil
}

func (x *LoginResponse) Validate() error {

	if err := x.validateAccessToken(); err != nil {
		return err
	}

	if err := x.validateRefreshToken(); err != nil {
		return err
	}

	return nil
}

func (x *LoginResponse) validateAccessToken() error {

	return nil
}

func (x *LoginResponse) validateRefreshToken() error {

	return nil
}

func (x *HelloRequest) Validate() error {

	if err := x.validateName(); err != nil {
		return err
	}

	return nil
}

func (x *HelloRequest) validateName() error {

	if genproto.IsNilOrEmpty(x.Name) {
		return &genproto.ValidationError{
			Key:     "Name",
			Message: "is required",
		}
	}

	if strings.TrimSpace(x.GetName()) == "" {
		return &genproto.ValidationError{
			Key:     "Name",
			Message: "can not be blank",
		}
	}

	if len(x.GetName()) < 2 {
		return &genproto.ValidationError{
			Key:     "Name",
			Message: fmt.Sprintf("must be at least %d characters long", 2),
		}
	}

	if len(x.GetName()) > 64 {
		return &genproto.ValidationError{
			Key:     "Name",
			Message: fmt.Sprintf("must be at most %d characters long", 64),
		}
	}

	return nil
}

func (x *HelloResponse) Validate() error {

	if err := x.validateMessage(); err != nil {
		return err
	}

	return nil
}

func (x *HelloResponse) validateMessage() error {

	return nil
}

func (x *GetImageRequest) Validate() error {

	if err := x.validateKey(); err != nil {
		return err
	}

	if err := x.validateWidth(); err != nil {
		return err
	}

	if err := x.validateHeight(); err != nil {
		return err
	}

	return nil
}

func (x *GetImageRequest) validateKey() error {

	if genproto.IsNilOrEmpty(x.Key) {
		return &genproto.ValidationError{
			Key:     "Key",
			Message: "is required",
		}
	}

	if strings.TrimSpace(x.GetKey()) == "" {
		return &genproto.ValidationError{
			Key:     "Key",
			Message: "can not be blank",
		}
	}

	if len(x.GetKey()) > 64 {
		return &genproto.ValidationError{
			Key:     "Key",
			Message: fmt.Sprintf("must be at most %d characters long", 64),
		}
	}

	return nil
}

func (x *GetImageRequest) validateWidth() error {

	if x.GetWidth() <= 0 {
		return &genproto.ValidationError{
			Key:     "Width",
			Message: fmt.Sprintf("must be greater than %d", 0),
		}
	}

	if x.GetWidth() > 1024 {
		return &genproto.ValidationError{
			Key:     "Width",
			Message: fmt.Sprintf("must be less than or equal to %d", 1024),
		}
	}

	return nil
}

func (x *GetImageRequest) validateHeight() error {

	if x.GetHeight() <= 0 {
		return &genproto.ValidationError{
			Key:     "Height",
			Message: fmt.Sprintf("must be greater than %d", 0),
		}
	}

	if x.GetHeight() > 1024 {
		return &genproto.ValidationError{
			Key:     "Height",
			Message: fmt.Sprintf("must be less than or equal to %d", 1024),
		}
	}

	return nil
}

func (x *GetImageResponse) Validate() error {

	if err := x.validateFile(); err != nil {
		return err
	}

	return nil
}

func (x *GetImageResponse) validateFile() error {

	return nil
}

func (x *UploadImageRequest) Validate() error {

	if err := x.validateImage(); err != nil {
		return err
	}

	if err := x.validateCategory(); err != nil {
		return err
	}

	return nil
}

func (x *UploadImageRequest) validateImage() error {

	if genproto.IsNilOrEmpty(x.Image) {
		return &genproto.ValidationError{
			Key:     "Image",
			Message: "is required",
		}
	}

	if len(x.GetImage()) < 1 {
		return &genproto.ValidationError{
			Key:     "Image",
			Message: fmt.Sprintf("must be at least %d bytes long", 1),
		}
	}

	if len(x.GetImage()) > 1048576 {
		return &genproto.ValidationError{
			Key:     "Image",
			Message: fmt.Sprintf("must be at most %d bytes long", 1048576),
		}
	}

	return nil
}

func (x *UploadImageRequest) validateCategory() error {

	if genproto.IsNilOrEmpty(x.Category) {
		return &genproto.ValidationError{
			Key:     "Category",
			Message: "is required",
		}
	}

	if strings.TrimSpace(x.GetCategory()) == "" {
		return &genproto.ValidationError{
			Key:     "Category",
			Message: "can not be blank",
		}
	}

	inList := []string{"avatar", "background", "post"}
	if !genproto.StringIn(x.GetCategory(), inList) {
		return &genproto.ValidationError{
			Key:     "Category",
			Message: fmt.Sprintf("must be one of %v", inList),
		}
	}

	return nil
}

func (x *UploadImageResponse) Validate() error {

	if err := x.validateKey(); err != nil {
		return err
	}

	return nil
}

func (x *UploadImageResponse) validateKey() error {

	return nil
}

func (x *UpdateProgressRequest) Validate() error {

	if err := x.validateProgress(); err != nil {
		return err
	}

	if err := x.validateCatIds(); err != nil {
		return err
	}

	return nil
}

func (x *UpdateProgressRequest) validateProgress() error {

	if genproto.IsNilOrEmpty(x.Progress) {
		return &genproto.ValidationError{
			Key:     "Progress",
			Message: "is required",
		}
	}

	if x.GetProgress() < 0 {
		return &genproto.ValidationError{
			Key:     "Progress",
			Message: fmt.Sprintf("must be greater than or equal to %f", 0),
		}
	}

	if x.GetProgress() > 1 {
		return &genproto.ValidationError{
			Key:     "Progress",
			Message: fmt.Sprintf("must be less than or equal to %f", 1),
		}
	}

	return nil
}

func (x *UpdateProgressRequest) validateCatIds() error {

	if genproto.IsNilOrEmpty(x.CatIds) {
		return &genproto.ValidationError{
			Key:     "CatIds",
			Message: "is required",
		}
	}

	if len(x.GetCatIds()) < 1 {
		return &genproto.ValidationError{
			Key:     "CatIds",
			Message: fmt.Sprintf("must have at least %d items", 1),
		}
	}

	if len(x.GetCatIds()) > 10 {
		return &genproto.ValidationError{
			Key:     "CatIds",
			Message: fmt.Sprintf("must have at most %d items", 10),
		}
	}

	return nil
}

func (x *UpdateProgressResponse) Validate() error {

	if err := x.validateProgress(); err != nil {
		return err
	}

	return nil
}

func (x *UpdateProgressResponse) validateProgress() error {

	return nil
}
