package logger

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap/zapcore"
)

func jsonFileEntry(t *testing.T, cfg Config) (Entry, string) {
	t.Helper()
	if cfg.FilePath == "" {
		cfg.FilePath = t.TempDir()
	}
	cfg.Console = false
	cfg.Encoding = JSONEncoding
	return NewEntry(cfg), cfg.FilePath
}

func syncLogger(t *testing.T, e Entry) {
	t.Helper()
	if err := e.Logger().Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate") {
		t.Fatalf("sync logger: %v", err)
	}
}

func readJSONLog(t *testing.T, dir, name string) []map[string]any {
	t.Helper()
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()

	var lines []map[string]any
	s := bufio.NewScanner(f)
	for s.Scan() {
		var line map[string]any
		if err := json.Unmarshal(s.Bytes(), &line); err != nil {
			t.Fatalf("decode %s line %q: %v", name, s.Text(), err)
		}
		lines = append(lines, line)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan %s: %v", name, err)
	}
	return lines
}

func TestStructuredLevelMethodsTrimFields(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{
		Level: DebugLevel,
		Trimmer: &TrimConfig{
			ArrLimit: 2,
			StrLimit: 4,
			Ignores:  []string{"secret"},
		},
	})

	e.Debugw("debug", "value", "abcdef")
	e.Infow("info", "items", []int{1, 2, 3}, "secret", "token")
	e.Warnw("warn", "ok", true)
	e.Errorw("error", "err", errors.New("failure"))
	syncLogger(t, e)

	lines := readJSONLog(t, dir, "debug.log")
	if len(lines) != 4 {
		t.Fatalf("debug.log line count = %d, want 4", len(lines))
	}
	if got := lines[0]["value"]; got != "abcd..." {
		t.Errorf("trimmed string = %#v, want %q", got, "abcd...")
	}
	items, ok := lines[1]["items"].([]any)
	if !ok || len(items) != 2 || items[0] != float64(1) || items[1] != float64(2) {
		t.Errorf("trimmed items = %#v, want [1 2]", lines[1]["items"])
	}
	if _, exists := lines[1]["secret"]; exists {
		t.Error("ignored field secret was logged")
	}
	if got := lines[3]["err"]; got != "fail..." {
		t.Errorf("trimmed error = %#v, want %q", got, "fail...")
	}
}

func TestPrintAndStructuredMethodsHaveDistinctSemantics(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{Level: InfoLevel})
	e.Info("hello", "world")
	e.Infow("structured", "key", "value")
	syncLogger(t, e)

	lines := readJSONLog(t, dir, "info.log")
	if got := lines[0]["msg"]; got != "helloworld" {
		t.Errorf("Info message = %#v, want print-style concatenation", got)
	}
	if _, exists := lines[0]["key"]; exists {
		t.Error("Info unexpectedly interpreted arguments as fields")
	}
	if got := lines[1]["key"]; got != "value" {
		t.Errorf("Infow key = %#v, want value", got)
	}
}

func TestWithDoesNotMutateParent(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{Level: InfoLevel})
	child := e.WithField("child", true).WithFields(Fields{"request_id": "r1"})
	e.Info("parent")
	child.Info("child")
	syncLogger(t, e)

	lines := readJSONLog(t, dir, "info.log")
	if _, exists := lines[0]["child"]; exists {
		t.Error("parent inherited a child field")
	}
	if lines[1]["child"] != true || lines[1]["request_id"] != "r1" {
		t.Errorf("child fields = %#v", lines[1])
	}
}

func TestSpecialFieldsUseTrimmer(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{
		Level: InfoLevel,
		Trimmer: &TrimConfig{
			StrLimit: 4,
			Ignores:  []string{"error"},
		},
	})
	e.WithError(errors.New("secret failure")).Info("without-error")
	e.WithRequestID("request-too-long").Info("with-request")
	syncLogger(t, e)

	lines := readJSONLog(t, dir, "info.log")
	if _, exists := lines[0]["error"]; exists {
		t.Error("WithError bypassed the ignored-fields configuration")
	}
	if got := lines[1]["request_id"]; got != "requ..." {
		t.Errorf("request_id = %#v, want %q", got, "requ...")
	}
}

func TestSetLevelAffectsExistingChildrenAndOutput(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{Level: DebugLevel})
	child := e.With("scope", "child")
	e.SetLevel(ErrorLevel)
	child.Info("suppressed")
	child.Error("visible")
	syncLogger(t, e)

	if got := child.GetLevel(); got != ErrorLevel {
		t.Fatalf("child level = %v, want ErrorLevel", got)
	}
	body, err := os.ReadFile(filepath.Join(dir, "debug.log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "suppressed") || !strings.Contains(string(body), "visible") {
		t.Errorf("unexpected debug.log contents: %s", body)
	}
}

func TestNoConfiguredSinkIsSilent(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false}).(*logEntry)
	if e.base.Desugar().Core().Enabled(zapcore.Level(DebugLevel)) {
		t.Fatal("logger with no configured sink should use a no-op core")
	}
	if !e.Enabled(DebugLevel) {
		t.Fatal("Enabled should still report the configured level threshold")
	}
}

func TestEntryAndPackageCaller(t *testing.T) {
	e, dir := jsonFileEntry(t, Config{Level: InfoLevel, Caller: true})
	_, file, entryLine, _ := runtime.Caller(0)
	e.Info("entry-caller") // the assertion below allows this line plus one
	syncLogger(t, e)

	line := readJSONLog(t, dir, "info.log")[0]
	caller, _ := line["caller"].(string)
	wantFile := filepath.Base(file)
	if !strings.Contains(caller, wantFile+":"+strconv.Itoa(entryLine+1)) {
		t.Errorf("entry caller = %q, want %s:%d", caller, wantFile, entryLine+1)
	}

	globalDir := t.TempDir()
	Configure(Config{Level: InfoLevel, Encoding: JSONEncoding, FilePath: globalDir, Caller: true})
	defer Configure(defaultConfig)
	_, file, packageLine, _ := runtime.Caller(0)
	Info("package-caller")
	syncLogger(t, L())

	line = readJSONLog(t, globalDir, "info.log")[0]
	caller, _ = line["caller"].(string)
	if !strings.Contains(caller, filepath.Base(file)+":"+strconv.Itoa(packageLine+1)) {
		t.Errorf("package caller = %q, want %s:%d", caller, filepath.Base(file), packageLine+1)
	}
}

func TestGlobalConfigureIsConcurrencySafe(t *testing.T) {
	defer Configure(defaultConfig)
	const iterations = 100
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				Configure(Config{Level: DebugLevel, Console: false})
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				Info("concurrent")
				_ = L()
				_ = GetLevel()
			}
		}()
	}
	wg.Wait()
}

func TestNewEntryWithOptsUsesDefaultsAndIgnoresNilOption(t *testing.T) {
	dir := t.TempDir()
	e := NewEntryWithOpts(nil, WithFilePath(dir), WithConsole(false), WithEncoding(JSONEncoding))
	if got := e.GetLevel(); got != InfoLevel {
		t.Fatalf("default level = %v, want InfoLevel", got)
	}
	e.Info("default-options")
	syncLogger(t, e)
	if lines := readJSONLog(t, dir, "info.log"); len(lines) != 1 {
		t.Fatalf("info.log line count = %d, want 1", len(lines))
	}
}
