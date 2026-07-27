package ginext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/annotation/v3"
	"github.com/tenz-io/gokit/ginext/v3/errcode"
)

// maxBodyBytes 是 [BindAndValidate] 默认允许读取的请求体上限(10 MiB)。
// 它用 [http.MaxBytesReader] 套在 c.Request.Body 外,使 io.ReadAll 超过
// 上限时返回错误而非无限读满内存(防止 OOM)。可用 [SetMaxBodyBytes]
// 在启动时覆盖,例如对大文件上传场景放宽上限。
//
// 对 multipart 请求,ParseMultipartForm 的内存/临时文件阈值
// (10<<20)与此值对齐,但它不是请求大小上限 —— MaxBytesReader 才是。
var maxBodyBytes int64 = 10 << 20

// SetMaxBodyBytes 覆盖 [BindAndValidate] 读取请求体的字节上限。在进程
// 启动时(任何请求到来前)调用一次。传入 0 或负值恢复默认(10 MiB)。
func SetMaxBodyBytes(n int64) {
	if n <= 0 {
		maxBodyBytes = 10 << 20
		return
	}
	maxBodyBytes = n
}

// BindAndValidate 把一个请求绑定到 ptr 所指 struct 并校验它。
//
// 每个字段的取值来源用 `bind` 标签声明,例如:
//
//	GET /path/:id?offset=1
//	type Req struct {
//	    ID     int64  `bind:"uri,name=id"   validate:"required,gt=0"`
//	    Offset int    `bind:"query,name=offset" default:"0" validate:"gte=0"`
//	    Auth   string `bind:"header,name=Authorization"`
//	    Title  string `bind:"form,name=title" validate:"required,min_len=1"`
//	    File   []byte `bind:"file,name=file" validate:"required"`
//	}
//
// 对 JSON 请求体(Content-Type: application/json),body 会被直接 unmarshal
// 到 ptr;字段名由 `json` 标签控制。
//
// 处理顺序:
//  1. 用 `default` 标签填充默认值;
//  2. 绑定 uri/query/header/multipart/form 等显式来源(它们一旦设置就具有
//     权威性,见下);
//  3. 若是 JSON 请求体,先快照步骤 2 已设置的显式来源字段的值,
//     再 unmarshal JSON(它会全量覆盖 struct),随后把快照值回填 —— 这防止
//     JSON body 覆盖路径/查询/头中声明的字段(例如 /user/1 + {"id":2}
//     不再让 ID 变成 2);
//  4. 运行 [annotation.Validate] 收集全部失败项。
//
// 任何失败都会被 [warpError] 包装成 400 [errcode.Error],以便响应层统一渲染。
//
// 请求体大小:[BindAndValidate] 用 [http.MaxBytesReader] 套住 c.Request.Body,
// 读取超过 [maxBodyBytes] 字节会以错误返回(而非无限读满内存)。默认 10 MiB,
// 用 [SetMaxBodyBytes] 覆盖。
func BindAndValidate(c *gin.Context, ptr any) (err error) {
	defer func() {
		if err != nil {
			err = warpError(err)
		}
	}()

	// 限制请求体大小,防止 JSON/form/multipart 绑定时无限读取致 OOM。
	// MaxBytesReader 在超限时返回 *http.MaxBytesError,Read 端据此返回 400。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	if err = annotation.ApplyDefaults(ptr); err != nil {
		return fmt.Errorf("parse default value: %w", err)
	}

	if has, e := tryBindURI(c, ptr); has && e != nil {
		return annotation.Err("uri", "", e.Error())
	}
	if has, e := tryBindQuery(c, ptr); has && e != nil {
		return annotation.Err("query", "", e.Error())
	}
	if has, e := tryBindHeader(c, ptr); has && e != nil {
		return annotation.Err("header", "", e.Error())
	}
	if has, e := tryBindMultipart(c, ptr); has && e != nil {
		return annotation.Err("multipart", "", e.Error())
	}
	if has, e := tryBindForm(c, ptr); has && e != nil {
		return annotation.Err("form", "", e.Error())
	}

	// JSON 请求体在显式来源之后处理,且不得覆盖它们。snapshotOverlayFields
	// 先记下显式来源字段的当前值,tryBindJSON 全量 unmarshal 后用 restoreOverlayFields
	// 回填,从而保证 URI/Query/Header/Form/File 字段的权威性。
	if isJSONRequest(c) {
		snap, snapErr := snapshotOverlayFields(ptr)
		if snapErr != nil {
			return snapErr
		}
		if has, e := tryBindJSON(c, ptr); has {
			if e != nil {
				return e
			}
			// 回填被 JSON 覆盖的显式来源字段。
			if re := restoreOverlayFields(ptr, snap); re != nil {
				return re
			}
		}
	}

	if err = annotation.Validate(ptr); err != nil {
		return err
	}
	return nil
}

// isJSONRequest 报告请求是否为 JSON 请求体(Content-Type 前缀 application/json)。
func isJSONRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json")
}

// overlayFieldSnapshot 记录一个显式来源(uri/query/header/form/file)字段
// 在 JSON unmarshal 前的值,供 unmarshal 后回填,防止 JSON 覆盖。
type overlayFieldSnapshot struct {
	field *annotation.Field
	value reflect.Value // 字段当前值的可设置拷贝
}

// snapshotOverlayFields 为 ptr 中声明了 uri/query/header 来源的字段快照
// 当前值。这些来源在 URL/头中,与 JSON body 共存(例如 POST /user/1 配
// JSON body),因此 JSON unmarshal 可能覆盖它们 —— 快照用于回填。
//
// form/file 来源**不**纳入:JSON 请求的 Content-Type 是 application/json,
// 与 x-www-form-urlencoded / multipart 互斥,二者不会同时绑定,故无覆盖风险;
// 而把 form 字段纳入快照反而会撤销 JSON 对同名 json-tag 字段的合法填充。
func snapshotOverlayFields(ptr any) ([]overlayFieldSnapshot, error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return nil, e
	}
	root, e := rootValue(ptr)
	if e != nil {
		return nil, e
	}
	var snap []overlayFieldSnapshot
	for _, f := range plan.FieldsBySource(annotation.BindURI) {
		snap = appendSnapshot(snap, root, f)
	}
	for _, f := range plan.FieldsBySource(annotation.BindQuery) {
		snap = appendSnapshot(snap, root, f)
	}
	for _, f := range plan.FieldsBySource(annotation.BindHeader) {
		snap = appendSnapshot(snap, root, f)
	}
	return snap, nil
}

// appendSnapshot 把 f 在 root 上的当前值拷贝进快照(深拷贝以隔离切片/指针)。
func appendSnapshot(snap []overlayFieldSnapshot, root reflect.Value, f *annotation.Field) []overlayFieldSnapshot {
	fv := root.FieldByIndex(f.Index)
	snap = append(snap, overlayFieldSnapshot{
		field: f,
		value: reflect.ValueOf(fv.Interface()), // 拷贝值
	})
	return snap
}

// restoreOverlayFields 把快照值回填到对应字段,撤销 JSON unmarshal 的覆盖。
// 仅当字段在快照中记录了值时才回填。
func restoreOverlayFields(ptr any, snap []overlayFieldSnapshot) error {
	root, e := rootValue(ptr)
	if e != nil {
		return e
	}
	for _, s := range snap {
		fv := root.FieldByIndex(s.field.Index)
		if fv.CanSet() {
			fv.Set(s.value)
		}
	}
	return nil
}

// Validate 对已填充的 ptr 运行 annotation 校验。它不会绑定任何来源,仅在
// 调用方已自行填充结构体后做一次校验。
func Validate(c *gin.Context, ptr any) (err error) {
	defer func() {
		if err != nil {
			err = warpError(err)
		}
	}()
	if err = annotation.Validate(ptr); err != nil {
		return err
	}
	return nil
}

// rootValue 解析 ptr 背后的 struct 值,用于按字段索引(field index)访问。
func rootValue(ptr any) (reflect.Value, error) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected a non-nil pointer to a struct")
	}
	return rv.Elem(), nil
}

// fieldByIndex 返回 f 在 root 上的可设置值。
func fieldByIndex(root reflect.Value, f *annotation.Field) reflect.Value {
	return root.FieldByIndex(f.Index)
}

// tryBindURI 绑定路由参数,例如 /path/:id -> ID int64 `bind:"uri,name=id"`。
func tryBindURI(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	uriFields := plan.FieldsBySource(annotation.BindURI)
	if len(uriFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range uriFields {
		val := c.Param(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindQuery 绑定 query 参数,例如 ?offset=1 -> Offset `bind:"query,name=offset"`。
func tryBindQuery(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	qFields := plan.FieldsBySource(annotation.BindQuery)
	if len(qFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range qFields {
		val := c.Query(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindHeader 绑定 header,例如 Authorization: Bearer ... -> Auth `bind:"header,name=Authorization"`。
func tryBindHeader(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	hFields := plan.FieldsBySource(annotation.BindHeader)
	if len(hFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range hFields {
		val := c.GetHeader(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindForm 绑定 application/x-www-form-urlencoded 请求体(仅 POST/PUT)。
func tryBindForm(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		return false, nil
	}
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.Err("method", "",
			fmt.Sprintf("invalid method %s for form request, should be POST or PUT", c.Request.Method))
	}

	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	formFields := plan.FieldsBySource(annotation.BindForm)
	if len(formFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range formFields {
		// 字段缺失(readFormAndSetField 返回 nil)才跳过;字段一旦出现,
		// 解析/类型错误就必须返回 400,而不是因非必填而吞掉(否则可选字段
		// age=abc 会被静默置零)。
		if e := readFormAndSetField(c, root, f); e != nil {
			return true, e
		}
	}
	return true, nil
}

// tryBindJSON 把 JSON 请求体 unmarshal 到 ptr(POST/PUT/PATCH)。
func tryBindJSON(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		return false, nil
	}
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodDelete {
		return false, nil
	}
	body, e := io.ReadAll(c.Request.Body)
	if e != nil {
		// 请求体超限(*http.MaxBytesError)直接原样返回,使 warpError 把它
		// 映射为 413,而非被包成 annotation 校验错误(那样会被误归 400)。
		var maxErr *http.MaxBytesError
		if errors.As(e, &maxErr) {
			return true, e
		}
		return true, annotation.Err("body", "", fmt.Sprintf("error reading request body: %s", e.Error()))
	}
	if e = json.Unmarshal(body, ptr); e != nil {
		return true, annotation.Err("json_format", "", fmt.Sprintf("error unmarshalling request body: %s", e.Error()))
	}
	return true, nil
}

// tryBindMultipart 绑定 multipart/form-data 请求体(文件 + 表单字段)。
func tryBindMultipart(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		return false, nil
	}
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.Err("method", "",
			fmt.Sprintf("invalid method %s for multipart request, should be POST or PUT", c.Request.Method))
	}

	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	fileFields := plan.FieldsBySource(annotation.BindFile)
	formFields := plan.FieldsBySource(annotation.BindForm)
	if len(fileFields) == 0 {
		return true, annotation.Err("multipart", "", "no file field found in struct")
	}

	if e = c.Request.ParseMultipartForm(10 << 20); e != nil {
		return true, annotation.Err("multipart", "", fmt.Sprintf("error parsing multipart form: %s", e.Error()))
	}

	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}

	for _, f := range fileFields {
		// 字段缺失(无对应 file part)时 readFileAndSetField 返回 nil;
		// 字段一旦出现,读取/类型错误必须返回 400。
		if e = readFileAndSetField(c, root, f); e != nil {
			return true, e
		}
	}
	for _, f := range formFields {
		if e = readFormAndSetField(c, root, f); e != nil {
			return true, e
		}
	}
	return true, nil
}

// readFileAndSetField 从 multipart 表单读取名为 f.BindName 的文件,写入 root。
func readFileAndSetField(c *gin.Context, root reflect.Value, f *annotation.Field) error {
	file, _, err := c.Request.FormFile(f.BindName)
	if err != nil {
		return annotation.Err(f.BindName, "", fmt.Sprintf("error getting file %s: %s", f.BindName, err.Error()))
	}
	defer func() { _ = file.Close() }()

	bs, err := io.ReadAll(file)
	if err != nil {
		return annotation.Err(f.BindName, "", fmt.Sprintf("error reading file %s: %s", f.BindName, err.Error()))
	}
	if len(bs) == 0 {
		return nil
	}
	if err := annotation.Set(fieldByIndex(root, f), bs); err != nil {
		return annotation.Err(f.BindName, "", err.Error())
	}
	return nil
}

// readFormAndSetField 从 multipart/urlencoded 表单读取 f.BindName 的文本值,写入 root。
func readFormAndSetField(c *gin.Context, root reflect.Value, f *annotation.Field) error {
	val := c.Request.FormValue(f.BindName)
	if val == "" {
		return nil
	}
	if err := annotation.SetString(fieldByIndex(root, f), val); err != nil {
		return annotation.Err(f.BindName, "", err.Error())
	}
	return nil
}

// warpError 把绑定/校验错误翻译成一个 HTTP 400 [errcode.Error],使响应层
// ([ErrorResponse])能渲染一个一致的 400 并带上字段级错误细节。
//
// 注意:[errcode.New] 的第一个参数是业务 code,HTTP status 缺省为 200。
// 这里显式把 status 也设为 [http.StatusBadRequest],使绑定/校验失败既在
// 响应体 code 上、也在 HTTP 状态码上都呈现 400 —— 这修复了 v2 warpError
// 仅设 code 不设 status、导致 HTTP 仍返回 200 的缺陷。
func warpError(err error) error {
	if err == nil {
		return nil
	}
	// 请求体超限:[http.MaxBytesReader] 在超限时返回 *http.MaxBytesError,
	// 映射为 413 而非 400。
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return errcode.New(http.StatusRequestEntityTooLarge,
			fmt.Sprintf("request body too large (limit %d bytes)", maxErr.Limit),
			http.StatusRequestEntityTooLarge)
	}
	// 收集到的校验失败:全部上报。
	if verrs, ok := annotation.AsErrors(err); ok && verrs.Has() {
		return errcode.BadRequest(http.StatusBadRequest, verrs.Error())
	}
	// 畸形 JSON 载荷。
	var jsonErr *json.UnmarshalTypeError
	if errors.As(err, &jsonErr) {
		return errcode.BadRequest(http.StatusBadRequest, jsonErr.Error())
	}
	return errcode.BadRequest(http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
}
