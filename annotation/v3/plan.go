package annotation

import (
	"reflect"
	"strings"
	"sync"
)

// BindSource names where a value should be read from when binding an HTTP
// request. It is parsed from the bind tag (e.g. bind:"uri,name=id" -> URI);
// the transport layer decides how to fetch each source.
type BindSource string

const (
	// BindNone means no bind source declared.
	BindNone BindSource = ""
	// BindURI binds from a route parameter.
	BindURI BindSource = "uri"
	// BindQuery binds from a URL query parameter.
	BindQuery BindSource = "query"
	// BindHeader binds from an HTTP header.
	BindHeader BindSource = "header"
	// BindForm binds from a form (urlencoded or multipart).
	BindForm BindSource = "form"
	// BindFile binds an uploaded file's bytes.
	BindFile BindSource = "file"
)

const (
	tagDefault  = "default"
	tagBind     = "bind"
	tagJSON     = "json"
	tagProto    = "protobuf"
	tagValidate = "validate"
)

// Field is one cached entry for a leaf or nested field in a struct plan.
type Field struct {
	// Path is the dotted path from the root struct, e.g. "Config.Addr.Street".
	Path string
	// Name is the Go field name.
	Name string
	// Index is the reflect field index chain usable with FieldByIndex.
	Index []int
	// Type is the field's reflect.Type.
	Type reflect.Type
	// StructField is the reflect.StructField (tags, exported flag, ...).
	StructField reflect.StructField

	// BindName is the resolved external name (bind name -> json -> protobuf ->
	// Go name).
	BindName string
	// BindSource is where the value should be read from, or BindNone.
	BindSource BindSource

	// Default is the default-tag value, or "" when none.
	Default string

	// rules are the compiled validation rules for this field, in declaration
	// order. nil for fields with no validate tag.
	rules []namedRule

	// children are nested struct fields, in declaration order. nil for leaves
	// (including slices/maps whose elements are not validated structurally).
	children []*Field
}

// IsRequired reports whether the field carries a "required" rule.
func (f *Field) IsRequired() bool {
	for _, r := range f.rules {
		if r.name == "required" {
			return true
		}
	}
	return false
}

// Plan is the cached structural description of a struct type. It is built once
// per type and reused across requests.
type Plan struct {
	rootType reflect.Type
	fields   []*Field
}

// Fields returns the cached top-level fields in declaration order.
func (p *Plan) Fields() []*Field { return p.fields }

// Walk visits every field (including nested structs) in declaration order. The
// callback may return false to stop the walk early.
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

// FieldsBySource returns top-level fields whose BindSource matches src, in
// declaration order. Convenience for transport binders.
func (p *Plan) FieldsBySource(src BindSource) []*Field {
	var out []*Field
	for _, f := range p.fields {
		if f.BindSource == src {
			out = append(out, f)
		}
	}
	return out
}

// PlanFor returns the cached Plan for the struct behind ptr, building it on
// first use. ptr must be a non-nil pointer to a struct.
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
	// Only root plans are cached. A nested field's Path depends on its parent,
	// so caching nested plans by type alone would leak the first-seen path into
	// every later use of that type.
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
		// A recursive pointer (for example, a linked-list Next field) is a
		// legitimate Go type. Stop expanding the static plan at the cycle rather
		// than publishing an incomplete plan or recurring forever.
		return &Plan{rootType: t}, nil
	}
	ancestors[t] = true
	defer delete(ancestors, t)

	p := &Plan{rootType: t}
	var fields []*Field
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		fpath := joinPath(parent, sf.Name)

		ft := sf.Type
		// Peel at most one pointer for structural inspection so *Nested still
		// recurses, but keep the original type on the Field for setters.
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

// peelPtr returns the struct type inside ft (peeling one pointer level), or nil
// if ft is not a (pointer to a) struct. time.Time and similar are treated as
// leaves because they expose their own validation semantics via rules, not via
// their internal fields.
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

// resolveBindName picks the external name for binding, in priority order:
// bind tag's name=... -> json tag's first element -> protobuf tag's name=...
// -> Go field name.
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
	// Preserve declaration order. Iterating the map returned by tagMap made a
	// malformed multi-source tag choose a different source nondeterministically.
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

// tagMap parses "k1=v1,k2=v2,k" style tags into a map. Bare keys (no '=') map
// to "".
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
