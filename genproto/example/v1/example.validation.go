// Code generated by protoc-gen-go-idl. DO NOT EDIT.
package v1

import (
	"context"
	"github.com/tenz-io/gokit/genproto"
	"github.com/tenz-io/gokit/genproto/go/custom/idl"
	"google.golang.org/protobuf/proto"
)

func init() {

	_LoginRequest := &LoginRequest{}
	genproto.Register("LoginRequest", _LoginRequest.ValidateRule())

	_LoginResponse := &LoginResponse{}
	genproto.Register("LoginResponse", _LoginResponse.ValidateRule())

	_HelloRequest := &HelloRequest{}
	genproto.Register("HelloRequest", _HelloRequest.ValidateRule())

	_HelloResponse := &HelloResponse{}
	genproto.Register("HelloResponse", _HelloResponse.ValidateRule())

	_GetImageRequest := &GetImageRequest{}
	genproto.Register("GetImageRequest", _GetImageRequest.ValidateRule())

	_GetImageResponse := &GetImageResponse{}
	genproto.Register("GetImageResponse", _GetImageResponse.ValidateRule())

	_UploadImageRequest := &UploadImageRequest{}
	genproto.Register("UploadImageRequest", _UploadImageRequest.ValidateRule())

	_UploadImageResponse := &UploadImageResponse{}
	genproto.Register("UploadImageResponse", _UploadImageResponse.ValidateRule())

}

func (x *LoginRequest) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *LoginRequest) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Username": &idl.Field{
			Str: &idl.StringField{
				Required: proto.Bool(true),
				NotBlank: proto.Bool(true),
				MinLen:   proto.Int64(2),
				MaxLen:   proto.Int64(64),
			},
		},

		"Password": &idl.Field{
			Str: &idl.StringField{
				Required: proto.Bool(true),
				NotBlank: proto.Bool(true),
				MinLen:   proto.Int64(6),
				MaxLen:   proto.Int64(64),
			},
		},
	}
}

func (x *LoginResponse) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *LoginResponse) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"AccessToken": &idl.Field{},

		"RefreshToken": &idl.Field{},
	}
}

func (x *HelloRequest) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *HelloRequest) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Name": &idl.Field{
			Str: &idl.StringField{
				Default:  proto.String("goer"),
				Required: proto.Bool(true),
				NotBlank: proto.Bool(true),
				MinLen:   proto.Int64(2),
				MaxLen:   proto.Int64(64),
			},
		},
	}
}

func (x *HelloResponse) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *HelloResponse) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Message": &idl.Field{},
	}
}

func (x *GetImageRequest) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *GetImageRequest) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Key": &idl.Field{
			Str: &idl.StringField{
				Required: proto.Bool(true),
				NotBlank: proto.Bool(true),
				MaxLen:   proto.Int64(64),
			},
		},

		"Width": &idl.Field{
			Int: &idl.IntField{
				Default: proto.Int64(640),
				Gt:      proto.Int64(0),
				Lte:     proto.Int64(1024),
			},
		},

		"Height": &idl.Field{
			Int: &idl.IntField{
				Default: proto.Int64(480),
				Gt:      proto.Int64(0),
				Lte:     proto.Int64(1024),
			},
		},
	}
}

func (x *GetImageResponse) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *GetImageResponse) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"File": &idl.Field{},
	}
}

func (x *UploadImageRequest) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *UploadImageRequest) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Image": &idl.Field{
			Bytes: &idl.BytesField{
				Required: proto.Bool(true),
				MinLen:   proto.Int64(1),
				MaxLen:   proto.Int64(1048576),
			},
		},

		"Category": &idl.Field{
			Str: &idl.StringField{
				Default:  proto.String("post"),
				Required: proto.Bool(true),
				NotBlank: proto.Bool(true),
				In: []string{
					"avatar",
					"background",
					"post",
				},
			},
		},
	}
}

func (x *UploadImageResponse) Validate(_ context.Context) error {
	return genproto.Validate(x)
}

func (x *UploadImageResponse) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{

		"Key": &idl.Field{},
	}
}
