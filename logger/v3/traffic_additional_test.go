package logger

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func trafficEntry(t *testing.T, trim *TrimConfig) (*logEntry, string) {
	t.Helper()
	dir := t.TempDir()
	e := NewEntry(Config{
		Level:       DebugLevel,
		Console:     false,
		Traffic:     true,
		TrafficPath: dir,
		Trimmer:     trim,
	}).(*logEntry)
	if e.traffic == nil {
		t.Fatal("traffic logger was not created")
	}
	return e, dir
}

func readTraffic(t *testing.T, e *logEntry, dir string) string {
	t.Helper()
	if err := e.traffic.Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate") {
		t.Fatalf("sync traffic logger: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "traffic.log"))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func trafficLineCount(body string) int {
	body = strings.TrimSpace(body)
	if body == "" {
		return 0
	}
	return strings.Count(body, "\n") + 1
}

func TestTrafficTrimsResponseFieldsAndWritesOnce(t *testing.T) {
	e, dir := trafficEntry(t, &TrimConfig{ArrLimit: 2, StrLimit: 4})
	rec := e.StartTraffic("command-too-long").WithTyp(TrafficTypSend)
	rec.End(
		map[string]any{"value": "response-too-long", "items": []int{1, 2, 3}},
		"200",
		"extra", "field-too-long",
	)
	rec.End(map[string]any{"duplicate": true}, "500")

	body := readTraffic(t, e, dir)
	if lines := trafficLineCount(body); lines != 1 {
		t.Fatalf("traffic line count = %d, want 1; body=%q", lines, body)
	}
	for _, want := range []string{"comm...", "resp...", "fiel...", "\"items\":[1,2]", "send"} {
		if !strings.Contains(body, want) {
			t.Errorf("traffic log missing %q: %s", want, body)
		}
	}
	for _, unwanted := range []string{"command-too-long", "response-too-long", "field-too-long", "duplicate"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("traffic log contains untrimmed/duplicate value %q: %s", unwanted, body)
		}
	}
}

func TestTrafficEndWithErrorAndNilMethods(t *testing.T) {
	e, dir := trafficEntry(t, nil)
	e.StartTraffic("failed").EndWithError(errors.New("timeout"), "attempt", 2)
	body := readTraffic(t, e, dir)
	for _, want := range []string{"failed", "error", "timeout", "attempt", "2"} {
		if !strings.Contains(body, want) {
			t.Errorf("traffic log missing %q: %s", want, body)
		}
	}

	var rec *TrafficRec
	if got := rec.WithTyp(TrafficTypSend); got != nil {
		t.Fatal("WithTyp on nil receiver should return nil")
	}
	rec.End(nil, "")
	rec.EndWithError(nil)
}

func TestTrafficWithTypAndEndAreConcurrencySafe(t *testing.T) {
	e, dir := trafficEntry(t, nil)
	rec := e.StartTraffic("concurrent")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				rec.WithTyp(TrafficTypSend)
			} else {
				rec.End(nil, "200")
			}
		}(i)
	}
	wg.Wait()

	body := readTraffic(t, e, dir)
	if lines := trafficLineCount(body); lines != 1 {
		t.Fatalf("traffic line count = %d, want 1; body=%q", lines, body)
	}
}

func TestTrafficDirectoryFailureDisablesTraffic(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := NewEntry(Config{
		Console:     false,
		Traffic:     true,
		TrafficPath: filepath.Join(file, "child"),
	})
	if rec := e.StartTraffic("disabled"); rec != nil {
		t.Fatal("traffic should be disabled when its directory cannot be created")
	}
}

func TestTrafficJSONEncodingAndIgnoredResponse(t *testing.T) {
	dir := t.TempDir()
	e := NewEntry(Config{
		Console:     false,
		Encoding:    JSONEncoding,
		Traffic:     true,
		TrafficPath: dir,
		Trimmer:     &TrimConfig{Ignores: []string{"resp"}},
	}).(*logEntry)
	e.StartTraffic("json").End(map[string]any{"secret": true}, "200")
	if err := e.traffic.Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate") {
		t.Fatal(err)
	}
	lines := readJSONLog(t, dir, "traffic.log")
	if len(lines) != 1 {
		t.Fatalf("traffic line count = %d, want 1", len(lines))
	}
	if _, exists := lines[0]["resp"]; exists {
		t.Error("ignored traffic response was logged")
	}
	if lines[0]["summary"] == nil || lines[0]["cmd"] != "json" || lines[0]["code"] != "200" {
		t.Errorf("unexpected traffic JSON: %#v", lines[0])
	}
}
