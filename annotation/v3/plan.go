package annotation

import (
	"reflect"
	"strings"
	"sync"
)

// BindSource 命名绑定 HTTP 请求时从何处读取值。它从 bind tag 解析
// (例如 bind:"uri,name=id" -> URI);由 transport 层决定如何取每个 source。
type BindSource string

const (
	// BindNone 表示未声明 bind source。
	BindNone BindSource = ""
	// BindURI 从路由参数绑定。
	BindURI BindSource = "uri"
	// BindQuery 从 URL query 参数绑定。
	BindQuery BindSource = "query"
	// BindHeader 从 HTTP header 绑定。
	BindHeader BindSource = "header"
	// BindForm 从 form(urlencoded 或 multipart)绑定。
	BindForm BindSource = "form"
	// BindFile 绑定上传文件的字节。
	BindFile BindSource = "file"
)

const (
	tagDefault  = "default"
	tagBind     = "bind"
	tagJSON     = "json"
	tagProto    = "protobuf"
	tagValidate = "validate"
)

// Field 是 struct plan 中叶子或嵌套 field 的一个缓存条目。
type Field struct {
	// Path 是从根 struct 起的点分路径,例如 "Config.Addr.Street"。
	Path string
	// Name 是 Go field 名。
	Name string
	// Index 是可用于 FieldByIndex 的 reflect field index 链。
	Index []int
	// Type 是 field 的 reflect.Type。
	Type reflect.Type
	// StructField 是 reflect.StructField(tag、导出标志、...)。
	StructField reflect.StructField

	// BindName 是解析后的外部名(bind name -> json -> protobuf -> Go 名)。
	BindName string
	// BindSource 是值应从何处读取,或 BindNone。
	BindSource BindSource

	// Default 是 default-tag 的值,无则为 ""。
	Default string

	// rules 是该 field 编译后的校验规则,按声明顺序排列。无 validate tag 的
	// field 为 nil。
	rules []namedRule

	// children 是嵌套 struct field,按声明顺序排列。叶子(包括元素不被结构性
	// 校验的 slice/map)为 nil。
	children []*Field
}

// IsRequired 报告该 field 是否带有 "required" 规则。
func (f *Field) IsRequired() bool {
	for _, r := range f.rules {
		if r.name == "required" {
			return true
		}
	}
	return false
}

// Plan 是某个 struct 类型的缓存结构描述。每个类型构建一次,跨请求复用。
type Plan struct {
	rootType reflect.Type
	fields   []*Field
}

// Fields 返回按声明顺序排列的缓存顶层 field。
func (p *Plan) Fields() []*Field { return p.fields }

// Walk 按声明顺序访问每个 field(包括嵌套 struct)。回调可返回 false 以提前终止遍历。
func (p *Plan) Walk(fn func(*Field) bool) {
	for _, f := range p.fields {
		if !walkField(f, fn) {
			return
		}
	}
}

func walkField(f *Field, fn func(*Field) bool) bool {
	if !fn(f) {
		return false
	}
	for _, c := range f.children {
		if !walkField(c, fn) {
			return false
		}
	}
	return true
}

// FieldsBySource 返回 BindSource 匹配 src 的顶层 field,按声明顺序排列。
// 为 transport binder 提供便利。
func (p *Plan) FieldsBySource(src BindSource) []*Field {
	var out []*Field
	for _, f := range p.fields {
		if f.BindSource == src {
			out = append(out, f)
		}
	}
	return out
}

// PlanFor 返回 ptr 所指 struct 的缓存 Plan,首次使用时构建。ptr 必须是指向
// struct 的非空 pointer。
func PlanFor(ptr any) (*Plan, error) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, &invalidArgError{"annotation.PlanFor: expected a non-nil pointer to a struct"}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil, &invalidArgError{"annotation.PlanFor: expected a pointer to a struct"}
	}
	return planForType(rv.Type(), "")
}

var planCache sync.Map // reflect.Type -> *Plan

func planForType(t reflect.Type, parent string) (*Plan, error) {
	// 仅缓存根 plan。嵌套 field 的 Path 依赖其 parent,因此仅按类型缓存嵌套
	// plan 会把首次见到的路径泄漏到该类型后续的每次使用中。
	if parent == "" {
		if cached, ok := planCache.Load(t); ok {
			return cached.(*Plan), nil
		}
	}

	p, err := buildPlan(t, parent, make(map[reflect.Type]bool))
	if err != nil {
		return nil, err
	}
	if parent != "" {
		return p, nil
	}
	actual, _ := planCache.LoadOrStore(t, p)
	return actual.(*Plan), nil
}

func buildPlan(t reflect.Type, parent string, ancestors map[reflect.Type]bool) (*Plan, error) {
	if ancestors[t] {
		// 递归 pointer(例如链表的 Next field)是合法的 Go 类型。在环处停止
		// 扩展静态 plan,而不是发布不完整的 plan 或无限递归。
		return &Plan{rootType: t}, nil
	}
	ancestors[t] = true
	defer delete(ancestors, t)

	p := &Plan{rootType: t}
	var fields []*Field
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported(未导出)
			continue
		}
		fpath := joinPath(parent, sf.Name)

		ft := sf.Type
		// 为结构性检查最多剥离一层 pointer,使 *Nested 仍可递归,但在 Field
		// 上保留原始类型以供 setter 使用。
		structType := peelPtr(ft)
		var children []*Field
		if structType != nil {
			cp, err := buildPlan(structType, fpath, ancestors)
			if err != nil {
				return nil, err
			}
			children = cp.fields
		}

		f := &Field{
			Path:        fpath,
			Name:        sf.Name,
			Index:       []int{i},
			Type:        ft,
			StructField: sf,
			BindName:    resolveBindName(sf),
			BindSource:  resolveBindSource(sf),
			Default:     sf.Tag.Get(tagDefault),
			children:    children,
		}
		if err := compileRules(f); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	p.fields = fields
	return p, nil
}

// peelPtr 返回 ft 内部的 struct 类型(剥离一层 pointer),若 ft 不是 struct
// 或指向 struct 的 pointer 则返回 nil。time.Time 及类似类型被视为叶子,因为
// 它们通过 rules 而非内部 field 暴露自身的校验语义。
func peelPtr(ft reflect.Type) reflect.Type {
	if ft == timeType {
		return nil
	}
	t := ft
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t
	}
	return nil
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "." + name
}

// resolveBindName 按优先级顺序选取绑定的外部名:
// bind tag 的 name=... -> json tag 的首元素 -> protobuf tag 的 name=... -> Go field 名。
func resolveBindName(sf reflect.StructField) string {
	if name, ok := bindName(sf.Tag); ok && name != "" {
		return name
	}
	if name, ok := jsonName(sf.Tag); ok && name != "" {
		return name
	}
	if name, ok := protoName(sf.Tag); ok && name != "" {
		return name
	}
	return sf.Name
}

func resolveBindSource(sf reflect.StructField) BindSource {
	// 保留声明顺序。遍历 tagMap 返回的 map 会使格式错误的多 source tag 不确定地
	// 选择不同 source。
	for _, elem := range strings.Split(sf.Tag.Get(tagBind), ",") {
		key, _, _ := strings.Cut(elem, "=")
		switch BindSource(key) {
		case BindURI, BindQuery, BindHeader, BindForm, BindFile:
			return BindSource(key)
		}
	}
	return BindNone
}

func bindName(tag reflect.StructTag) (string, bool) {
	m := tagMap(tag, tagBind)
	return m["name"], m["name"] != "" || hasKey(m, "name")
}

func hasKey(m map[string]string, k string) bool {
	_, ok := m[k]
	return ok
}

func jsonName(tag reflect.StructTag) (string, bool) {
	v := tag.Get(tagJSON)
	if v == "" {
		return "", false
	}
	parts := strings.SplitN(v, ",", 2)
	return parts[0], parts[0] != "-"
}

func protoName(tag reflect.StructTag) (string, bool) {
	return tagMap(tag, tagProto)["name"], true
}

// tagMap 将 "k1=v1,k2=v2,k" 风格的 tag 解析为 map。无 '=' 的裸键映射为 ""。
func tagMap(tag reflect.StructTag, name string) map[string]string {
	m := map[string]string{}
	v := tag.Get(name)
	if v == "" {
		return m
	}
	for _, elem := range strings.Split(v, ",") {
		if elem == "" {
			continue
		}
		if k, val, ok := strings.Cut(elem, "="); ok {
			m[k] = val
		} else {
			m[elem] = ""
		}
	}
	return m
}

type invalidArgError struct{ msg string }

func (e *invalidArgError) Error() string { return e.msg }
