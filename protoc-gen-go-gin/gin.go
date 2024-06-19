package main

import (
	"github.com/tenz-io/gokit/ginext"
	"net/http"
	"regexp"
	"strings"

	"github.com/tenz-io/gokit/genproto/go/custom/common"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPkg         = protogen.GoImportPath("context")
	ginPkg             = protogen.GoImportPath("github.com/gin-gonic/gin")
	metadataPkg        = protogen.GoImportPath("github.com/tenz-io/gokit/ginext/metadata")
	ginextPkg          = protogen.GoImportPath("github.com/tenz-io/gokit/ginext")
	deprecationComment = "// Deprecated: Do not use."
)

var methodSets = make(map[string]int)

// generateFile generates a _gin.pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_gin.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by github.com/tenz-io/gokit/protoc-gen-go-gin. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the github.com/tenz-io/gokit/protoc-gen-go-gin package it is being compiled against.")
	g.P("// ", contextPkg.Ident(""), metadataPkg.Ident(""))
	g.P("// ", ginPkg.Ident(""), ginextPkg.Ident(""))
	g.P()

	for _, service := range file.Services {
		genService(gen, file, g, service)
	}
	return g
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, s *protogen.Service) {
	if s.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &service{
		Name:     s.GoName,
		FullName: string(s.Desc.FullName()),
		FilePath: file.Desc.Path(),
	}

	for _, m := range s.Methods {
		sd.Methods = append(sd.Methods, genMethod(m)...)
	}
	g.P(sd.execute())
}

func genMethod(m *protogen.Method) []*method {
	var (
		methods  []*method
		role     int32
		authType int32
	)

	if auth, ok := proto.GetExtension(m.Desc.Options(), common.E_Auth).(*common.Auth); ok {
		role = int32(auth.GetRole())
		authType = int32(auth.GetType())
	}

	rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
	if rule != nil && ok {
		for _, bind := range rule.AdditionalBindings {
			methods = append(methods, buildHTTPRule(m, bind, role, authType))
		}
		methods = append(methods, buildHTTPRule(m, rule, role, authType))
		return methods
	}

	// default method
	methods = append(methods, defaultMethod(m))
	return methods
}

// defaultMethodPath 根据函数名生成 http 路由
// 例如: GetBlogArticles ==> get: /blog/articles
// 如果方法名首个单词不是 http method 映射，那么默认返回 POST
func defaultMethod(m *protogen.Method) *method {
	names := strings.Split(toSnakeCase(m.GoName), "_")
	var (
		paths      []string
		httpMethod string
		path       string
	)

	switch strings.ToUpper(names[0]) {
	case http.MethodGet, "FIND", "QUERY", "LIST", "SEARCH":
		httpMethod = http.MethodGet
	case http.MethodPost, "CREATE":
		httpMethod = http.MethodPost
	case http.MethodPut, "UPDATE":
		httpMethod = http.MethodPut
	case http.MethodPatch:
		httpMethod = http.MethodPatch
	case http.MethodDelete:
		httpMethod = http.MethodDelete
	default:
		httpMethod = "POST"
		paths = names
	}

	if len(paths) > 0 {
		path = strings.Join(paths, "/")
	}

	if len(names) > 1 {
		path = strings.Join(names[1:], "/")
	}

	md := buildMethodDesc(m, httpMethod, path, ginext.RoleAnonymous, ginext.AuthTypeWeb)
	md.Body = "*"
	return md
}

func buildHTTPRule(m *protogen.Method, rule *annotations.HttpRule, role, authType int32) *method {
	var (
		path       string
		httpMethod string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		httpMethod = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		httpMethod = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		httpMethod = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		httpMethod = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		httpMethod = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		httpMethod = pattern.Custom.Kind
	}
	md := buildMethodDesc(m, httpMethod, path, role, authType)
	return md
}

func buildMethodDesc(m *protogen.Method, httpMethod, path string, role, authType int32) *method {
	defer func() { methodSets[m.GoName]++ }()
	md := &method{
		Name:     m.GoName,
		Num:      methodSets[m.GoName],
		Request:  m.Input.GoIdent.GoName,
		Reply:    m.Output.GoIdent.GoName,
		Path:     path,
		Method:   httpMethod,
		Role:     role,
		AuthType: authType,
	}
	md.initPathParams()
	return md
}

var matchFirstCap = regexp.MustCompile("([A-Z])([A-Z][a-z])")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(input string) string {
	output := matchFirstCap.ReplaceAllString(input, "${1}_${2}")
	output = matchAllCap.ReplaceAllString(output, "${1}_${2}")
	output = strings.ReplaceAll(output, "-", "_")
	return strings.ToLower(output)
}
