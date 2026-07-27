package annotation

import (
	"errors"
	"reflect"
)

// Validate 对 ptr 所指 struct 的每个 field 运行每条规则,并返回所有失败。
// ptr 必须是指向 struct 的非空 pointer。nil 表示无失败。
func Validate(ptr any) error {
	p, err := PlanFor(ptr)
	if err != nil {
		return err
	}
	root := reflect.ValueOf(ptr).Elem()
	var errs ValidationErrors
	validateField(p.fields, root, &errs)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// QuickValidate 运行与 Validate 相同的规则,但在遇到首个失败时即返回单个
// FieldError。nil 表示无失败。当你只需知道输入是否有效(以及为何无效的一个
// 原因),而不必付出收集所有错误的代价时使用。
func QuickValidate(ptr any) error {
	p, err := PlanFor(ptr)
	if err != nil {
		return err
	}
	root := reflect.ValueOf(ptr).Elem()
	if fe, ok := validateFieldFirst(p.fields, root); ok {
		return fe
	}
	return nil
}

func validateField(fields []*Field, parent reflect.Value, errs *ValidationErrors) {
	for _, f := range fields {
		fv := parent.FieldByIndex(f.Index)
		// 对原始(可能为 nil)的值运行该 field 自身的规则,使 "required"
		// 能捕获 nil 的嵌套 pointer。
		for _, r := range f.rules {
			ok, msg := r.run(fv)
			if ok {
				continue
			}
			fe := FieldError{Field: f.Path, Rule: r.name, Param: r.param}
			if r.msg != "" {
				fe.Msg = r.msg
			} else {
				fe.Msg = msg
			}
			*errs = append(*errs, fe)
		}

		// 仅当存在时递归进入嵌套 struct。此处不实例化 nil 的嵌套 pointer
		// (那会掩盖 "required" 的失败);ApplyDefaults 负责接通可选 default。
		if len(f.children) == 0 {
			continue
		}
		cur := fv
		if cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				continue
			}
			cur = cur.Elem()
		}
		if cur.Kind() == reflect.Struct && cur.IsValid() && cur.CanInterface() {
			validateField(f.children, cur, errs)
		}
	}
}

// validateFieldFirst 是 validateField 的短路孪生:在首个失败规则处停止,并以
// 单个 FieldError 返回。
func validateFieldFirst(fields []*Field, parent reflect.Value) (FieldError, bool) {
	for _, f := range fields {
		fv := parent.FieldByIndex(f.Index)
		for _, r := range f.rules {
			if ok, msg := r.run(fv); !ok {
				fe := FieldError{Field: f.Path, Rule: r.name, Param: r.param}
				if r.msg != "" {
					fe.Msg = r.msg
				} else {
					fe.Msg = msg
				}
				return fe, true
			}
		}
		if len(f.children) == 0 {
			continue
		}
		cur := fv
		if cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				continue
			}
			cur = cur.Elem()
		}
		if cur.Kind() == reflect.Struct && cur.IsValid() && cur.CanInterface() {
			if fe, ok := validateFieldFirst(f.children, cur); ok {
				return fe, true
			}
		}
	}
	return FieldError{}, false
}

// AsErrors 从 Validate 返回的 error 中提取收集到的失败。当 error 不来自
// validator 时 ok 为 false。
func AsErrors(err error) (errs ValidationErrors, ok bool) {
	if err == nil {
		return nil, false
	}
	if ve, ok := err.(ValidationErrors); ok {
		return ve, true
	}
	var target ValidationErrors
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}
