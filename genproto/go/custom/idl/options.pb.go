// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: custom/idl/options.proto

// custom interface definition language

package idl

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Method_AuthRole int32

const (
	Method_ANONYMOUS Method_AuthRole = 0 // anonymous, no need to verify
	Method_ADMIN     Method_AuthRole = 1 // admin
	Method_USER      Method_AuthRole = 2 // user
)

// Enum value maps for Method_AuthRole.
var (
	Method_AuthRole_name = map[int32]string{
		0: "ANONYMOUS",
		1: "ADMIN",
		2: "USER",
	}
	Method_AuthRole_value = map[string]int32{
		"ANONYMOUS": 0,
		"ADMIN":     1,
		"USER":      2,
	}
)

func (x Method_AuthRole) Enum() *Method_AuthRole {
	p := new(Method_AuthRole)
	*p = x
	return p
}

func (x Method_AuthRole) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Method_AuthRole) Descriptor() protoreflect.EnumDescriptor {
	return file_custom_idl_options_proto_enumTypes[0].Descriptor()
}

func (Method_AuthRole) Type() protoreflect.EnumType {
	return &file_custom_idl_options_proto_enumTypes[0]
}

func (x Method_AuthRole) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Method_AuthRole.Descriptor instead.
func (Method_AuthRole) EnumDescriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{0, 0}
}

type Method_AuthType int32

const (
	Method_WEB  Method_AuthType = 0 // web page, verify by cookie
	Method_REST Method_AuthType = 1 // restful api, verify by token
)

// Enum value maps for Method_AuthType.
var (
	Method_AuthType_name = map[int32]string{
		0: "WEB",
		1: "REST",
	}
	Method_AuthType_value = map[string]int32{
		"WEB":  0,
		"REST": 1,
	}
)

func (x Method_AuthType) Enum() *Method_AuthType {
	p := new(Method_AuthType)
	*p = x
	return p
}

func (x Method_AuthType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Method_AuthType) Descriptor() protoreflect.EnumDescriptor {
	return file_custom_idl_options_proto_enumTypes[1].Descriptor()
}

func (Method_AuthType) Type() protoreflect.EnumType {
	return &file_custom_idl_options_proto_enumTypes[1]
}

func (x Method_AuthType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Method_AuthType.Descriptor instead.
func (Method_AuthType) EnumDescriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{0, 1}
}

type Method struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role Method_AuthRole `protobuf:"varint,1,opt,name=role,proto3,enum=custom.idl.Method_AuthRole" json:"role,omitempty"`
	Type Method_AuthType `protobuf:"varint,2,opt,name=type,proto3,enum=custom.idl.Method_AuthType" json:"type,omitempty"`
	// Types that are assignable to Route:
	//
	//	*Method_Get
	//	*Method_Put
	//	*Method_Post
	//	*Method_Delete
	//	*Method_Patch
	//	*Method_Head
	//	*Method_Options
	Route isMethod_Route `protobuf_oneof:"route"`
}

func (x *Method) Reset() {
	*x = Method{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Method) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Method) ProtoMessage() {}

func (x *Method) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Method.ProtoReflect.Descriptor instead.
func (*Method) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{0}
}

func (x *Method) GetRole() Method_AuthRole {
	if x != nil {
		return x.Role
	}
	return Method_ANONYMOUS
}

func (x *Method) GetType() Method_AuthType {
	if x != nil {
		return x.Type
	}
	return Method_WEB
}

func (m *Method) GetRoute() isMethod_Route {
	if m != nil {
		return m.Route
	}
	return nil
}

func (x *Method) GetGet() string {
	if x, ok := x.GetRoute().(*Method_Get); ok {
		return x.Get
	}
	return ""
}

func (x *Method) GetPut() string {
	if x, ok := x.GetRoute().(*Method_Put); ok {
		return x.Put
	}
	return ""
}

func (x *Method) GetPost() string {
	if x, ok := x.GetRoute().(*Method_Post); ok {
		return x.Post
	}
	return ""
}

func (x *Method) GetDelete() string {
	if x, ok := x.GetRoute().(*Method_Delete); ok {
		return x.Delete
	}
	return ""
}

func (x *Method) GetPatch() string {
	if x, ok := x.GetRoute().(*Method_Patch); ok {
		return x.Patch
	}
	return ""
}

func (x *Method) GetHead() string {
	if x, ok := x.GetRoute().(*Method_Head); ok {
		return x.Head
	}
	return ""
}

func (x *Method) GetOptions() string {
	if x, ok := x.GetRoute().(*Method_Options); ok {
		return x.Options
	}
	return ""
}

type isMethod_Route interface {
	isMethod_Route()
}

type Method_Get struct {
	Get string `protobuf:"bytes,21,opt,name=get,proto3,oneof"`
}

type Method_Put struct {
	Put string `protobuf:"bytes,22,opt,name=put,proto3,oneof"`
}

type Method_Post struct {
	Post string `protobuf:"bytes,23,opt,name=post,proto3,oneof"`
}

type Method_Delete struct {
	Delete string `protobuf:"bytes,24,opt,name=delete,proto3,oneof"`
}

type Method_Patch struct {
	Patch string `protobuf:"bytes,25,opt,name=patch,proto3,oneof"`
}

type Method_Head struct {
	Head string `protobuf:"bytes,26,opt,name=head,proto3,oneof"`
}

type Method_Options struct {
	Options string `protobuf:"bytes,27,opt,name=options,proto3,oneof"`
}

func (*Method_Get) isMethod_Route() {}

func (*Method_Put) isMethod_Route() {}

func (*Method_Post) isMethod_Route() {}

func (*Method_Delete) isMethod_Route() {}

func (*Method_Patch) isMethod_Route() {}

func (*Method_Head) isMethod_Route() {}

func (*Method_Options) isMethod_Route() {}

type Field struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Bind:
	//
	//	*Field_Header
	//	*Field_Uri
	//	*Field_Query
	//	*Field_Form
	//	*Field_File
	Bind  isField_Bind `protobuf_oneof:"bind"`
	Int   *IntField    `protobuf:"bytes,11,opt,name=int,proto3" json:"int,omitempty"`
	Str   *StringField `protobuf:"bytes,12,opt,name=str,proto3" json:"str,omitempty"`
	Bytes *BytesField  `protobuf:"bytes,13,opt,name=bytes,proto3" json:"bytes,omitempty"`
	Array *ArrayField  `protobuf:"bytes,14,opt,name=array,proto3" json:"array,omitempty"`
}

func (x *Field) Reset() {
	*x = Field{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Field) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Field) ProtoMessage() {}

func (x *Field) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Field.ProtoReflect.Descriptor instead.
func (*Field) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{1}
}

func (m *Field) GetBind() isField_Bind {
	if m != nil {
		return m.Bind
	}
	return nil
}

func (x *Field) GetHeader() string {
	if x, ok := x.GetBind().(*Field_Header); ok {
		return x.Header
	}
	return ""
}

func (x *Field) GetUri() string {
	if x, ok := x.GetBind().(*Field_Uri); ok {
		return x.Uri
	}
	return ""
}

func (x *Field) GetQuery() string {
	if x, ok := x.GetBind().(*Field_Query); ok {
		return x.Query
	}
	return ""
}

func (x *Field) GetForm() string {
	if x, ok := x.GetBind().(*Field_Form); ok {
		return x.Form
	}
	return ""
}

func (x *Field) GetFile() string {
	if x, ok := x.GetBind().(*Field_File); ok {
		return x.File
	}
	return ""
}

func (x *Field) GetInt() *IntField {
	if x != nil {
		return x.Int
	}
	return nil
}

func (x *Field) GetStr() *StringField {
	if x != nil {
		return x.Str
	}
	return nil
}

func (x *Field) GetBytes() *BytesField {
	if x != nil {
		return x.Bytes
	}
	return nil
}

func (x *Field) GetArray() *ArrayField {
	if x != nil {
		return x.Array
	}
	return nil
}

type isField_Bind interface {
	isField_Bind()
}

type Field_Header struct {
	Header string `protobuf:"bytes,1,opt,name=header,proto3,oneof"` // bind http header
}

type Field_Uri struct {
	Uri string `protobuf:"bytes,2,opt,name=uri,proto3,oneof"` // bind uri path
}

type Field_Query struct {
	Query string `protobuf:"bytes,3,opt,name=query,proto3,oneof"` // bind query string
}

type Field_Form struct {
	Form string `protobuf:"bytes,4,opt,name=form,proto3,oneof"` // bind form data
}

type Field_File struct {
	File string `protobuf:"bytes,5,opt,name=file,proto3,oneof"` // bind file data
}

func (*Field_Header) isField_Bind() {}

func (*Field_Uri) isField_Bind() {}

func (*Field_Query) isField_Bind() {}

func (*Field_Form) isField_Bind() {}

func (*Field_File) isField_Bind() {}

type IntField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Default  int64   `protobuf:"varint,1,opt,name=default,proto3" json:"default,omitempty"`   // default value
	Required bool    `protobuf:"varint,2,opt,name=required,proto3" json:"required,omitempty"` // required, can not be nil
	Gt       int64   `protobuf:"varint,3,opt,name=gt,proto3" json:"gt,omitempty"`             // greater than
	Gte      int64   `protobuf:"varint,4,opt,name=gte,proto3" json:"gte,omitempty"`           // greater than or equal
	Lt       int64   `protobuf:"varint,5,opt,name=lt,proto3" json:"lt,omitempty"`             // less than
	Lte      int64   `protobuf:"varint,6,opt,name=lte,proto3" json:"lte,omitempty"`           // less than or equal
	Eq       int64   `protobuf:"varint,7,opt,name=eq,proto3" json:"eq,omitempty"`             // equal
	Ne       int64   `protobuf:"varint,8,opt,name=ne,proto3" json:"ne,omitempty"`             // not equal
	In       []int64 `protobuf:"varint,9,rep,packed,name=in,proto3" json:"in,omitempty"`      // in list
}

func (x *IntField) Reset() {
	*x = IntField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IntField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IntField) ProtoMessage() {}

func (x *IntField) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IntField.ProtoReflect.Descriptor instead.
func (*IntField) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{2}
}

func (x *IntField) GetDefault() int64 {
	if x != nil {
		return x.Default
	}
	return 0
}

func (x *IntField) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *IntField) GetGt() int64 {
	if x != nil {
		return x.Gt
	}
	return 0
}

func (x *IntField) GetGte() int64 {
	if x != nil {
		return x.Gte
	}
	return 0
}

func (x *IntField) GetLt() int64 {
	if x != nil {
		return x.Lt
	}
	return 0
}

func (x *IntField) GetLte() int64 {
	if x != nil {
		return x.Lte
	}
	return 0
}

func (x *IntField) GetEq() int64 {
	if x != nil {
		return x.Eq
	}
	return 0
}

func (x *IntField) GetNe() int64 {
	if x != nil {
		return x.Ne
	}
	return 0
}

func (x *IntField) GetIn() []int64 {
	if x != nil {
		return x.In
	}
	return nil
}

type StringField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Default  string   `protobuf:"bytes,1,opt,name=default,proto3" json:"default,omitempty"`                    // default value
	Required bool     `protobuf:"varint,2,opt,name=required,proto3" json:"required,omitempty"`                 // required, can not be nil or empty
	NonBlank bool     `protobuf:"varint,3,opt,name=non_blank,json=nonBlank,proto3" json:"non_blank,omitempty"` // non blank, can not be blank string, such as " ", "\t", "\n"
	MinLen   int64    `protobuf:"varint,4,opt,name=min_len,json=minLen,proto3" json:"min_len,omitempty"`       // min length
	MaxLen   int64    `protobuf:"varint,5,opt,name=max_len,json=maxLen,proto3" json:"max_len,omitempty"`       // max length
	In       []string `protobuf:"bytes,6,rep,name=in,proto3" json:"in,omitempty"`                              // in list
	Pattern  string   `protobuf:"bytes,7,opt,name=pattern,proto3" json:"pattern,omitempty"`                    // regex pattern
}

func (x *StringField) Reset() {
	*x = StringField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringField) ProtoMessage() {}

func (x *StringField) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringField.ProtoReflect.Descriptor instead.
func (*StringField) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{3}
}

func (x *StringField) GetDefault() string {
	if x != nil {
		return x.Default
	}
	return ""
}

func (x *StringField) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *StringField) GetNonBlank() bool {
	if x != nil {
		return x.NonBlank
	}
	return false
}

func (x *StringField) GetMinLen() int64 {
	if x != nil {
		return x.MinLen
	}
	return 0
}

func (x *StringField) GetMaxLen() int64 {
	if x != nil {
		return x.MaxLen
	}
	return 0
}

func (x *StringField) GetIn() []string {
	if x != nil {
		return x.In
	}
	return nil
}

func (x *StringField) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

type BytesField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Default  []byte `protobuf:"bytes,1,opt,name=default,proto3" json:"default,omitempty"`              // default value
	Required bool   `protobuf:"varint,2,opt,name=required,proto3" json:"required,omitempty"`           // required, can not be nil or empty
	MinLen   int64  `protobuf:"varint,3,opt,name=min_len,json=minLen,proto3" json:"min_len,omitempty"` // min length
	MaxLen   int64  `protobuf:"varint,4,opt,name=max_len,json=maxLen,proto3" json:"max_len,omitempty"` // max length
}

func (x *BytesField) Reset() {
	*x = BytesField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesField) ProtoMessage() {}

func (x *BytesField) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesField.ProtoReflect.Descriptor instead.
func (*BytesField) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{4}
}

func (x *BytesField) GetDefault() []byte {
	if x != nil {
		return x.Default
	}
	return nil
}

func (x *BytesField) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *BytesField) GetMinLen() int64 {
	if x != nil {
		return x.MinLen
	}
	return 0
}

func (x *BytesField) GetMaxLen() int64 {
	if x != nil {
		return x.MaxLen
	}
	return 0
}

type ArrayField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Required bool       `protobuf:"varint,1,opt,name=required,proto3" json:"required,omitempty"`                 // required, can not be nil or empty
	MinItems int64      `protobuf:"varint,2,opt,name=min_items,json=minItems,proto3" json:"min_items,omitempty"` // min items in array
	MaxItems int64      `protobuf:"varint,3,opt,name=max_items,json=maxItems,proto3" json:"max_items,omitempty"` // max items in array
	Len      int64      `protobuf:"varint,4,opt,name=len,proto3" json:"len,omitempty"`                           // length of array
	Item     *ItemField `protobuf:"bytes,5,opt,name=item,proto3" json:"item,omitempty"`                          // item validation
}

func (x *ArrayField) Reset() {
	*x = ArrayField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArrayField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArrayField) ProtoMessage() {}

func (x *ArrayField) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArrayField.ProtoReflect.Descriptor instead.
func (*ArrayField) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{5}
}

func (x *ArrayField) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *ArrayField) GetMinItems() int64 {
	if x != nil {
		return x.MinItems
	}
	return 0
}

func (x *ArrayField) GetMaxItems() int64 {
	if x != nil {
		return x.MaxItems
	}
	return 0
}

func (x *ArrayField) GetLen() int64 {
	if x != nil {
		return x.Len
	}
	return 0
}

func (x *ArrayField) GetItem() *ItemField {
	if x != nil {
		return x.Item
	}
	return nil
}

type ItemField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Int   *IntField    `protobuf:"bytes,1,opt,name=int,proto3" json:"int,omitempty"`
	Str   *StringField `protobuf:"bytes,2,opt,name=str,proto3" json:"str,omitempty"`
	Bytes *BytesField  `protobuf:"bytes,3,opt,name=bytes,proto3" json:"bytes,omitempty"`
}

func (x *ItemField) Reset() {
	*x = ItemField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_custom_idl_options_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ItemField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ItemField) ProtoMessage() {}

func (x *ItemField) ProtoReflect() protoreflect.Message {
	mi := &file_custom_idl_options_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ItemField.ProtoReflect.Descriptor instead.
func (*ItemField) Descriptor() ([]byte, []int) {
	return file_custom_idl_options_proto_rawDescGZIP(), []int{6}
}

func (x *ItemField) GetInt() *IntField {
	if x != nil {
		return x.Int
	}
	return nil
}

func (x *ItemField) GetStr() *StringField {
	if x != nil {
		return x.Str
	}
	return nil
}

func (x *ItemField) GetBytes() *BytesField {
	if x != nil {
		return x.Bytes
	}
	return nil
}

var file_custom_idl_options_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*Method)(nil),
		Field:         50001,
		Name:          "custom.idl.method",
		Tag:           "bytes,50001,opt,name=method",
		Filename:      "custom/idl/options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*Field)(nil),
		Field:         60001,
		Name:          "custom.idl.field",
		Tag:           "bytes,60001,opt,name=field",
		Filename:      "custom/idl/options.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional custom.idl.Method method = 50001;
	E_Method = &file_custom_idl_options_proto_extTypes[0]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional custom.idl.Field field = 60001;
	E_Field = &file_custom_idl_options_proto_extTypes[1]
)

var File_custom_idl_options_proto protoreflect.FileDescriptor

var file_custom_idl_options_proto_rawDesc = []byte{
	0x0a, 0x18, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2f, 0x69, 0x64, 0x6c, 0x2f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe4, 0x02, 0x0a, 0x06, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x12, 0x2f, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1b, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x12, 0x2f, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e,
	0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x03, 0x67, 0x65, 0x74, 0x18, 0x15, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x67, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x03, 0x70, 0x75, 0x74,
	0x18, 0x16, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x70, 0x75, 0x74, 0x12, 0x14, 0x0a,
	0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x70,
	0x6f, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x18, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x16, 0x0a,
	0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x18, 0x19, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05,
	0x70, 0x61, 0x74, 0x63, 0x68, 0x12, 0x14, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x1a, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x1a, 0x0a, 0x07, 0x6f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x2e, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x52,
	0x6f, 0x6c, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x41, 0x4e, 0x4f, 0x4e, 0x59, 0x4d, 0x4f, 0x55, 0x53,
	0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x44, 0x4d, 0x49, 0x4e, 0x10, 0x01, 0x12, 0x08, 0x0a,
	0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x02, 0x22, 0x1d, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x57, 0x45, 0x42, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04,
	0x52, 0x45, 0x53, 0x54, 0x10, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x22,
	0xb0, 0x02, 0x0a, 0x05, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x18, 0x0a, 0x06, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x16, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x14, 0x0a, 0x04, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x04, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x14, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x26, 0x0a, 0x03, 0x69,
	0x6e, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f,
	0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x49, 0x6e, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x03,
	0x69, 0x6e, 0x74, 0x12, 0x29, 0x0a, 0x03, 0x73, 0x74, 0x72, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x03, 0x73, 0x74, 0x72, 0x12, 0x2c,
	0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x12, 0x2c, 0x0a, 0x05,
	0x61, 0x72, 0x72, 0x61, 0x79, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x41, 0x72, 0x72, 0x61, 0x79, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x52, 0x05, 0x61, 0x72, 0x72, 0x61, 0x79, 0x42, 0x06, 0x0a, 0x04, 0x62, 0x69,
	0x6e, 0x64, 0x22, 0xb4, 0x01, 0x0a, 0x08, 0x49, 0x6e, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x71,
	0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x71,
	0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x67, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x02, 0x67, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x03, 0x67, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x6c, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x02, 0x6c, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x74, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6c, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x65, 0x71, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02, 0x6e, 0x65, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x6e, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x6e, 0x18,
	0x09, 0x20, 0x03, 0x28, 0x03, 0x52, 0x02, 0x69, 0x6e, 0x22, 0xbc, 0x01, 0x0a, 0x0b, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x65, 0x66,
	0x61, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x64, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12,
	0x1b, 0x0a, 0x09, 0x6e, 0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x61, 0x6e, 0x6b, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x6e, 0x6f, 0x6e, 0x42, 0x6c, 0x61, 0x6e, 0x6b, 0x12, 0x17, 0x0a, 0x07,
	0x6d, 0x69, 0x6e, 0x5f, 0x6c, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6d,
	0x69, 0x6e, 0x4c, 0x65, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x78, 0x5f, 0x6c, 0x65, 0x6e,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x6e, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x02, 0x69, 0x6e, 0x12, 0x18,
	0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x22, 0x74, 0x0a, 0x0a, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x17, 0x0a, 0x07,
	0x6d, 0x69, 0x6e, 0x5f, 0x6c, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6d,
	0x69, 0x6e, 0x4c, 0x65, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x78, 0x5f, 0x6c, 0x65, 0x6e,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x22, 0x9f,
	0x01, 0x0a, 0x0a, 0x41, 0x72, 0x72, 0x61, 0x79, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x69, 0x6e,
	0x5f, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x6d, 0x69,
	0x6e, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x61, 0x78, 0x5f, 0x69, 0x74,
	0x65, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x6d, 0x61, 0x78, 0x49, 0x74,
	0x65, 0x6d, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x03, 0x6c, 0x65, 0x6e, 0x12, 0x29, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c,
	0x2e, 0x49, 0x74, 0x65, 0x6d, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d,
	0x22, 0x8c, 0x01, 0x0a, 0x09, 0x49, 0x74, 0x65, 0x6d, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x26,
	0x0a, 0x03, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x49, 0x6e, 0x74, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x52, 0x03, 0x69, 0x6e, 0x74, 0x12, 0x29, 0x0a, 0x03, 0x73, 0x74, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c,
	0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x03, 0x73, 0x74,
	0x72, 0x12, 0x2c, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x3a,
	0x4c, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd1, 0x86, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x3a, 0x48, 0x0a,
	0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe1, 0xd4, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2e, 0x69, 0x64, 0x6c, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x52, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65, 0x6e, 0x7a, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6f,
	0x6b, 0x69, 0x74, 0x2f, 0x67, 0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2f, 0x69, 0x64, 0x6c, 0x3b, 0x69, 0x64, 0x6c, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_custom_idl_options_proto_rawDescOnce sync.Once
	file_custom_idl_options_proto_rawDescData = file_custom_idl_options_proto_rawDesc
)

func file_custom_idl_options_proto_rawDescGZIP() []byte {
	file_custom_idl_options_proto_rawDescOnce.Do(func() {
		file_custom_idl_options_proto_rawDescData = protoimpl.X.CompressGZIP(file_custom_idl_options_proto_rawDescData)
	})
	return file_custom_idl_options_proto_rawDescData
}

var file_custom_idl_options_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_custom_idl_options_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_custom_idl_options_proto_goTypes = []interface{}{
	(Method_AuthRole)(0),               // 0: custom.idl.Method.AuthRole
	(Method_AuthType)(0),               // 1: custom.idl.Method.AuthType
	(*Method)(nil),                     // 2: custom.idl.Method
	(*Field)(nil),                      // 3: custom.idl.Field
	(*IntField)(nil),                   // 4: custom.idl.IntField
	(*StringField)(nil),                // 5: custom.idl.StringField
	(*BytesField)(nil),                 // 6: custom.idl.BytesField
	(*ArrayField)(nil),                 // 7: custom.idl.ArrayField
	(*ItemField)(nil),                  // 8: custom.idl.ItemField
	(*descriptorpb.MethodOptions)(nil), // 9: google.protobuf.MethodOptions
	(*descriptorpb.FieldOptions)(nil),  // 10: google.protobuf.FieldOptions
}
var file_custom_idl_options_proto_depIdxs = []int32{
	0,  // 0: custom.idl.Method.role:type_name -> custom.idl.Method.AuthRole
	1,  // 1: custom.idl.Method.type:type_name -> custom.idl.Method.AuthType
	4,  // 2: custom.idl.Field.int:type_name -> custom.idl.IntField
	5,  // 3: custom.idl.Field.str:type_name -> custom.idl.StringField
	6,  // 4: custom.idl.Field.bytes:type_name -> custom.idl.BytesField
	7,  // 5: custom.idl.Field.array:type_name -> custom.idl.ArrayField
	8,  // 6: custom.idl.ArrayField.item:type_name -> custom.idl.ItemField
	4,  // 7: custom.idl.ItemField.int:type_name -> custom.idl.IntField
	5,  // 8: custom.idl.ItemField.str:type_name -> custom.idl.StringField
	6,  // 9: custom.idl.ItemField.bytes:type_name -> custom.idl.BytesField
	9,  // 10: custom.idl.method:extendee -> google.protobuf.MethodOptions
	10, // 11: custom.idl.field:extendee -> google.protobuf.FieldOptions
	2,  // 12: custom.idl.method:type_name -> custom.idl.Method
	3,  // 13: custom.idl.field:type_name -> custom.idl.Field
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	12, // [12:14] is the sub-list for extension type_name
	10, // [10:12] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_custom_idl_options_proto_init() }
func file_custom_idl_options_proto_init() {
	if File_custom_idl_options_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_custom_idl_options_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Method); i {
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
		file_custom_idl_options_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Field); i {
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
		file_custom_idl_options_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IntField); i {
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
		file_custom_idl_options_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringField); i {
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
		file_custom_idl_options_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesField); i {
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
		file_custom_idl_options_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArrayField); i {
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
		file_custom_idl_options_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ItemField); i {
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
	file_custom_idl_options_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Method_Get)(nil),
		(*Method_Put)(nil),
		(*Method_Post)(nil),
		(*Method_Delete)(nil),
		(*Method_Patch)(nil),
		(*Method_Head)(nil),
		(*Method_Options)(nil),
	}
	file_custom_idl_options_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Field_Header)(nil),
		(*Field_Uri)(nil),
		(*Field_Query)(nil),
		(*Field_Form)(nil),
		(*Field_File)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_custom_idl_options_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_custom_idl_options_proto_goTypes,
		DependencyIndexes: file_custom_idl_options_proto_depIdxs,
		EnumInfos:         file_custom_idl_options_proto_enumTypes,
		MessageInfos:      file_custom_idl_options_proto_msgTypes,
		ExtensionInfos:    file_custom_idl_options_proto_extTypes,
	}.Build()
	File_custom_idl_options_proto = out.File
	file_custom_idl_options_proto_rawDesc = nil
	file_custom_idl_options_proto_goTypes = nil
	file_custom_idl_options_proto_depIdxs = nil
}
