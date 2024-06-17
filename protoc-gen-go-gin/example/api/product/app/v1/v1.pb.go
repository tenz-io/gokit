// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: product/app/v1/v1.proto

package v1

import (
	_ "go/custom/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetArticlesReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: bind:"uri,name=author_id" validate:"required,gt=0"
	AuthorId int32 `protobuf:"varint,1,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty" bind:"uri,name=author_id" validate:"required,gt=0"`
	// @inject_tag: bind:"query,name=title" validate:"max_len=200"
	Title string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty" bind:"query,name=title" validate:"max_len=200"`
	// @inject_tag: bind:"query,name=page" validate:"gt=0" default:"1"
	Page int32 `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty" bind:"query,name=page" validate:"gt=0" default:"1"`
	// @inject_tag: bind:"query,name=page_size" validate:"gt=0,lte=100" default:"20"
	PageSize int32 `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty" bind:"query,name=page_size" validate:"gt=0,lte=100" default:"20"`
}

func (x *GetArticlesReq) Reset() {
	*x = GetArticlesReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetArticlesReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetArticlesReq) ProtoMessage() {}

func (x *GetArticlesReq) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetArticlesReq.ProtoReflect.Descriptor instead.
func (*GetArticlesReq) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{0}
}

func (x *GetArticlesReq) GetAuthorId() int32 {
	if x != nil {
		return x.AuthorId
	}
	return 0
}

func (x *GetArticlesReq) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *GetArticlesReq) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *GetArticlesReq) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type GetArticlesResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total    int64      `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Articles []*Article `protobuf:"bytes,2,rep,name=articles,proto3" json:"articles,omitempty"`
}

func (x *GetArticlesResp) Reset() {
	*x = GetArticlesResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetArticlesResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetArticlesResp) ProtoMessage() {}

func (x *GetArticlesResp) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetArticlesResp.ProtoReflect.Descriptor instead.
func (*GetArticlesResp) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{1}
}

func (x *GetArticlesResp) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *GetArticlesResp) GetArticles() []*Article {
	if x != nil {
		return x.Articles
	}
	return nil
}

type CreateArticleReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: bind:"uri,name=author_id" validate:"required,gt=0"
	AuthorId int32 `protobuf:"varint,1,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty" bind:"uri,name=author_id" validate:"required,gt=0"`
	// @inject_tag: bind:"header,name=Authorization" validate:"required"
	Authorization string `protobuf:"bytes,2,opt,name=authorization,proto3" json:"authorization,omitempty" bind:"header,name=Authorization" validate:"required"`
	// @inject_tag: bind:"form,name=title" validate:"required,min_len=1,max_len=400"
	Title string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty" bind:"form,name=title" validate:"required,min_len=1,max_len=400"`
	// @inject_tag: bind:"form,name=content" validate:"required,min_len=1,max_len=100000"
	Content string `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty" bind:"form,name=content" validate:"required,min_len=1,max_len=100000"`
}

func (x *CreateArticleReq) Reset() {
	*x = CreateArticleReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateArticleReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArticleReq) ProtoMessage() {}

func (x *CreateArticleReq) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArticleReq.ProtoReflect.Descriptor instead.
func (*CreateArticleReq) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{2}
}

func (x *CreateArticleReq) GetAuthorId() int32 {
	if x != nil {
		return x.AuthorId
	}
	return 0
}

func (x *CreateArticleReq) GetAuthorization() string {
	if x != nil {
		return x.Authorization
	}
	return ""
}

func (x *CreateArticleReq) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *CreateArticleReq) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type CreateArticleResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ArticleId int32  `protobuf:"varint,1,opt,name=article_id,json=articleId,proto3" json:"article_id,omitempty"`
	Title     string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Content   string `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *CreateArticleResp) Reset() {
	*x = CreateArticleResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateArticleResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArticleResp) ProtoMessage() {}

func (x *CreateArticleResp) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArticleResp.ProtoReflect.Descriptor instead.
func (*CreateArticleResp) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{3}
}

func (x *CreateArticleResp) GetArticleId() int32 {
	if x != nil {
		return x.ArticleId
	}
	return 0
}

func (x *CreateArticleResp) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *CreateArticleResp) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type Article struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ArticleId int32  `protobuf:"varint,1,opt,name=article_id,json=articleId,proto3" json:"article_id,omitempty"`
	AuthorId  int32  `protobuf:"varint,2,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty"`
	Title     string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Content   string `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *Article) Reset() {
	*x = Article{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Article) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Article) ProtoMessage() {}

func (x *Article) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Article.ProtoReflect.Descriptor instead.
func (*Article) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{4}
}

func (x *Article) GetArticleId() int32 {
	if x != nil {
		return x.ArticleId
	}
	return 0
}

func (x *Article) GetAuthorId() int32 {
	if x != nil {
		return x.AuthorId
	}
	return 0
}

func (x *Article) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Article) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type UploadImageReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: bind:"uri,name=key" validate:"required,max_len=128,pattern=#abc123"
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty" bind:"uri,name=key" validate:"required,max_len=128,pattern=#abc123"`
	// @inject_tag: bind:"query,name=region" validate:"required,non_blank,len=2,pattern=#abc" default:"sg"
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty" bind:"query,name=region" validate:"required,non_blank,len=2,pattern=#abc" default:"sg"`
	// @inject_tag: bind:"header,name=Authorization" validate:"required"
	Authorization string `protobuf:"bytes,3,opt,name=authorization,proto3" json:"authorization,omitempty" bind:"header,name=Authorization" validate:"required"`
	// @inject_tag: bind:"file,name=image" validate:"min_len=1,max_len=102400"
	Image []byte `protobuf:"bytes,4,opt,name=image,proto3" json:"image,omitempty" bind:"file,name=image" validate:"min_len=1,max_len=102400"`
	// @inject_tag: bind:"form,name=filename"
	Filename string `protobuf:"bytes,5,opt,name=filename,proto3" json:"filename,omitempty" bind:"form,name=filename"`
}

func (x *UploadImageReq) Reset() {
	*x = UploadImageReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadImageReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadImageReq) ProtoMessage() {}

func (x *UploadImageReq) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadImageReq.ProtoReflect.Descriptor instead.
func (*UploadImageReq) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{5}
}

func (x *UploadImageReq) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *UploadImageReq) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *UploadImageReq) GetAuthorization() string {
	if x != nil {
		return x.Authorization
	}
	return ""
}

func (x *UploadImageReq) GetImage() []byte {
	if x != nil {
		return x.Image
	}
	return nil
}

func (x *UploadImageReq) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

type UploadImageResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *UploadImageResp) Reset() {
	*x = UploadImageResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadImageResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadImageResp) ProtoMessage() {}

func (x *UploadImageResp) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadImageResp.ProtoReflect.Descriptor instead.
func (*UploadImageResp) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{6}
}

func (x *UploadImageResp) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type GetImageReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: bind:"uri,name=key" validate:"required,max_len=128,pattern=#abc123"
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty" bind:"uri,name=key" validate:"required,max_len=128,pattern=#abc123"`
	// @inject_tag: bind:"query,name=region" validate:"required,non_blank,len=2,pattern=#abc" default:"sg"
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty" bind:"query,name=region" validate:"required,non_blank,len=2,pattern=#abc" default:"sg"`
}

func (x *GetImageReq) Reset() {
	*x = GetImageReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetImageReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetImageReq) ProtoMessage() {}

func (x *GetImageReq) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetImageReq.ProtoReflect.Descriptor instead.
func (*GetImageReq) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{7}
}

func (x *GetImageReq) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *GetImageReq) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

type GetImageResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	File []byte `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *GetImageResp) Reset() {
	*x = GetImageResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_product_app_v1_v1_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetImageResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetImageResp) ProtoMessage() {}

func (x *GetImageResp) ProtoReflect() protoreflect.Message {
	mi := &file_product_app_v1_v1_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetImageResp.ProtoReflect.Descriptor instead.
func (*GetImageResp) Descriptor() ([]byte, []int) {
	return file_product_app_v1_v1_proto_rawDescGZIP(), []int{8}
}

func (x *GetImageResp) GetFile() []byte {
	if x != nil {
		return x.File
	}
	return nil
}

var File_product_app_v1_v1_proto protoreflect.FileDescriptor

var file_product_app_v1_v1_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x76, 0x31,
	0x2f, 0x76, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2f,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x74, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69,
	0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61, 0x75, 0x74, 0x68,
	0x6f, 0x72, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61,
	0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1b,
	0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x5c, 0x0a, 0x0f, 0x47,
	0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x74,
	0x6f, 0x74, 0x61, 0x6c, 0x12, 0x33, 0x0a, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x2e, 0x61, 0x70, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52,
	0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x22, 0x85, 0x01, 0x0a, 0x10, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x12, 0x1b,
	0x0a, 0x09, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0d, 0x61,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x22, 0x62, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63,
	0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x61, 0x72, 0x74, 0x69,
	0x63, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x75, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x49, 0x64, 0x12,
	0x1b, 0x0a, 0x09, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x92, 0x01, 0x0a,
	0x0e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x24, 0x0a, 0x0d, 0x61, 0x75, 0x74,
	0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x69, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x22, 0x23, 0x0a, 0x0f, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x37, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x22,
	0x22, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x12, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x66,
	0x69, 0x6c, 0x65, 0x32, 0xd4, 0x03, 0x0a, 0x0b, 0x42, 0x6c, 0x6f, 0x67, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x7b, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x73, 0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x1a, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x22, 0x2b, 0x88, 0xb5, 0x18, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x21, 0x12,
	0x1f, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x2f, 0x7b, 0x61, 0x75, 0x74,
	0x68, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73,
	0x12, 0x7d, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x12, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65,
	0x52, 0x65, 0x71, 0x1a, 0x21, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70,
	0x70, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63,
	0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x22, 0x27, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x21, 0x22, 0x1f,
	0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x2f, 0x7b, 0x61, 0x75, 0x74, 0x68,
	0x6f, 0x72, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12,
	0x68, 0x0a, 0x0b, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x1e,
	0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x76, 0x31, 0x2e,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x1f,
	0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x76, 0x31, 0x2e,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x22,
	0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x22, 0x10, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6d, 0x61,
	0x67, 0x65, 0x73, 0x2f, 0x7b, 0x6b, 0x65, 0x79, 0x7d, 0x12, 0x5f, 0x0a, 0x08, 0x47, 0x65, 0x74,
	0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e,
	0x61, 0x70, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x71, 0x1a, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x70,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x73, 0x2f, 0x7b, 0x6b, 0x65, 0x79, 0x7d, 0x42, 0x1f, 0x5a, 0x1d, 0x65, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_product_app_v1_v1_proto_rawDescOnce sync.Once
	file_product_app_v1_v1_proto_rawDescData = file_product_app_v1_v1_proto_rawDesc
)

func file_product_app_v1_v1_proto_rawDescGZIP() []byte {
	file_product_app_v1_v1_proto_rawDescOnce.Do(func() {
		file_product_app_v1_v1_proto_rawDescData = protoimpl.X.CompressGZIP(file_product_app_v1_v1_proto_rawDescData)
	})
	return file_product_app_v1_v1_proto_rawDescData
}

var file_product_app_v1_v1_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_product_app_v1_v1_proto_goTypes = []interface{}{
	(*GetArticlesReq)(nil),    // 0: product.app.v1.GetArticlesReq
	(*GetArticlesResp)(nil),   // 1: product.app.v1.GetArticlesResp
	(*CreateArticleReq)(nil),  // 2: product.app.v1.CreateArticleReq
	(*CreateArticleResp)(nil), // 3: product.app.v1.CreateArticleResp
	(*Article)(nil),           // 4: product.app.v1.Article
	(*UploadImageReq)(nil),    // 5: product.app.v1.UploadImageReq
	(*UploadImageResp)(nil),   // 6: product.app.v1.UploadImageResp
	(*GetImageReq)(nil),       // 7: product.app.v1.GetImageReq
	(*GetImageResp)(nil),      // 8: product.app.v1.GetImageResp
}
var file_product_app_v1_v1_proto_depIdxs = []int32{
	4, // 0: product.app.v1.GetArticlesResp.articles:type_name -> product.app.v1.Article
	0, // 1: product.app.v1.BlogService.GetArticles:input_type -> product.app.v1.GetArticlesReq
	2, // 2: product.app.v1.BlogService.CreateArticle:input_type -> product.app.v1.CreateArticleReq
	5, // 3: product.app.v1.BlogService.UploadImage:input_type -> product.app.v1.UploadImageReq
	7, // 4: product.app.v1.BlogService.GetImage:input_type -> product.app.v1.GetImageReq
	1, // 5: product.app.v1.BlogService.GetArticles:output_type -> product.app.v1.GetArticlesResp
	3, // 6: product.app.v1.BlogService.CreateArticle:output_type -> product.app.v1.CreateArticleResp
	6, // 7: product.app.v1.BlogService.UploadImage:output_type -> product.app.v1.UploadImageResp
	8, // 8: product.app.v1.BlogService.GetImage:output_type -> product.app.v1.GetImageResp
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_product_app_v1_v1_proto_init() }
func file_product_app_v1_v1_proto_init() {
	if File_product_app_v1_v1_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_product_app_v1_v1_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetArticlesReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetArticlesResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateArticleReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateArticleResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Article); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadImageReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadImageResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetImageReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_product_app_v1_v1_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetImageResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_product_app_v1_v1_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_product_app_v1_v1_proto_goTypes,
		DependencyIndexes: file_product_app_v1_v1_proto_depIdxs,
		MessageInfos:      file_product_app_v1_v1_proto_msgTypes,
	}.Build()
	File_product_app_v1_v1_proto = out.File
	file_product_app_v1_v1_proto_rawDesc = nil
	file_product_app_v1_v1_proto_goTypes = nil
	file_product_app_v1_v1_proto_depIdxs = nil
}
