package plugin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestExpandTimeMacros(t *testing.T) {
	tr := backend.TimeRange{
		From: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
	}

	query := "SELECT * FROM events WHERE timestamp >= $__timeFrom AND timestamp < $__timeTo"
	result := expandTimeMacros(query, tr)

	expected := "SELECT * FROM events WHERE timestamp >= '2024-01-15 10:30:00' AND timestamp < '2024-01-16 10:30:00'"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestExpandTimeMacrosNoMacros(t *testing.T) {
	tr := backend.TimeRange{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	query := "SELECT count() FROM events"
	result := expandTimeMacros(query, tr)

	if result != query {
		t.Errorf("query without macros should be unchanged, got: %s", result)
	}
}

func TestNormalizeClickHouseType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"DateTime", "datetime"},
		{"Nullable(Float64)", "float64"},
		{"Nullable(String)", "string"},
		{"Int32", "int32"},
		{"UInt64", "uint64"},
		{"Nullable(DateTime)", "datetime"},
		{"Bool", "bool"},
	}

	for _, tt := range tests {
		result := normalizeClickHouseType(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeClickHouseType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestHogqlResponseToFrame(t *testing.T) {
	resp := &HogQLAPIResponse{
		Columns: []string{"timestamp", "count", "name"},
		Types:   []string{"DateTime", "Int64", "String"},
		Results: [][]interface{}{
			{"2024-01-15 10:30:00", float64(42), "page_view"},
			{"2024-01-15 11:00:00", float64(17), "click"},
		},
	}

	frame, err := hogqlResponseToFrame(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frame.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(frame.Fields))
	}

	if frame.Fields[0].Name != "timestamp" {
		t.Errorf("expected field name 'timestamp', got %q", frame.Fields[0].Name)
	}

	if frame.Fields[1].Name != "count" {
		t.Errorf("expected field name 'count', got %q", frame.Fields[1].Name)
	}

	if frame.Fields[2].Name != "name" {
		t.Errorf("expected field name 'name', got %q", frame.Fields[2].Name)
	}

	// Check row count
	if frame.Fields[0].Len() != 2 {
		t.Errorf("expected 2 rows, got %d", frame.Fields[0].Len())
	}

	// Verify numeric value
	val := frame.Fields[1].At(0)
	if fp, ok := val.(*float64); ok {
		if *fp != 42 {
			t.Errorf("expected count=42, got %f", *fp)
		}
	} else {
		t.Errorf("expected *float64, got %T", val)
	}

	// Verify string value
	strVal := frame.Fields[2].At(0)
	if sp, ok := strVal.(*string); ok {
		if *sp != "page_view" {
			t.Errorf("expected name='page_view', got %q", *sp)
		}
	} else {
		t.Errorf("expected *string, got %T", strVal)
	}
}

func TestHogqlResponseToFrameEmpty(t *testing.T) {
	resp := &HogQLAPIResponse{
		Columns: []string{},
		Types:   []string{},
		Results: [][]interface{}{},
	}

	frame, err := hogqlResponseToFrame(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frame.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(frame.Fields))
	}
}

func TestHogQLAPIResponseUnmarshalTypes(t *testing.T) {
	// New format: array of arrays [["ClickHouseType", "PostHogType"]]
	newFormat := `{"columns":["1"],"types":[["UInt8","UInt8"]],"results":[[1]]}`
	var resp1 HogQLAPIResponse
	if err := json.Unmarshal([]byte(newFormat), &resp1); err != nil {
		t.Fatalf("failed to unmarshal new format: %v", err)
	}
	if len(resp1.Types) != 1 || resp1.Types[0] != "UInt8" {
		t.Errorf("new format: expected [UInt8], got %v", resp1.Types)
	}

	// Old format: array of strings ["String"]
	oldFormat := `{"columns":["name"],"types":["String"],"results":[["test"]]}`
	var resp2 HogQLAPIResponse
	if err := json.Unmarshal([]byte(oldFormat), &resp2); err != nil {
		t.Fatalf("failed to unmarshal old format: %v", err)
	}
	if len(resp2.Types) != 1 || resp2.Types[0] != "String" {
		t.Errorf("old format: expected [String], got %v", resp2.Types)
	}

	// Multi-column new format
	multiCol := `{"columns":["ts","count"],"types":[["DateTime","DateTime"],["Int64","Int64"]],"results":[]}`
	var resp3 HogQLAPIResponse
	if err := json.Unmarshal([]byte(multiCol), &resp3); err != nil {
		t.Fatalf("failed to unmarshal multi-column: %v", err)
	}
	if len(resp3.Types) != 2 || resp3.Types[0] != "DateTime" || resp3.Types[1] != "Int64" {
		t.Errorf("multi-column: expected [DateTime, Int64], got %v", resp3.Types)
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		ok       bool
	}{
		{float64(42), 42, true},
		{float32(3.14), float64(float32(3.14)), true},
		{int(10), 10, true},
		{int64(100), 100, true},
		{"123.45", 123.45, true},
		{"not a number", 0, false},
		{nil, 0, false},
	}

	for _, tt := range tests {
		result, ok := toFloat64(tt.input)
		if ok != tt.ok {
			t.Errorf("toFloat64(%v) ok=%v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if ok && result != tt.expected {
			t.Errorf("toFloat64(%v) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}
