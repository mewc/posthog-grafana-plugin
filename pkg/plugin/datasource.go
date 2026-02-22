package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var (
	_ backend.QueryDataHandler   = (*PostHogDatasource)(nil)
	_ backend.CheckHealthHandler = (*PostHogDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*PostHogDatasource)(nil)
)

type PostHogDatasource struct {
	client *PostHogClient
}

func NewPostHogDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	pluginSettings, secrets, err := LoadSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	client := NewPostHogClient(pluginSettings.PostHogURL, pluginSettings.ProjectID, secrets.APIKey)
	return &PostHogDatasource{client: client}, nil
}

func (d *PostHogDatasource) Dispose() {}

func (d *PostHogDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.executeQuery(ctx, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *PostHogDatasource) executeQuery(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var qm PostHogQuery
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("failed to unmarshal query: %v", err))
	}

	if qm.RawHogQL == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "empty HogQL query")
	}

	hogql := expandTimeMacros(qm.RawHogQL, query.TimeRange)

	log.DefaultLogger.Debug("Executing HogQL query", "query", hogql)

	apiResp, err := d.client.ExecuteHogQL(ctx, hogql)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("query execution failed: %v", err))
	}

	frame, err := hogqlResponseToFrame(apiResp)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to convert response: %v", err))
	}

	return backend.DataResponse{Frames: data.Frames{frame}}
}

func expandTimeMacros(query string, timeRange backend.TimeRange) string {
	const layout = "2006-01-02 15:04:05"
	query = strings.ReplaceAll(query, "$__timeFrom", fmt.Sprintf("'%s'", timeRange.From.UTC().Format(layout)))
	query = strings.ReplaceAll(query, "$__timeTo", fmt.Sprintf("'%s'", timeRange.To.UTC().Format(layout)))
	return query
}

func hogqlResponseToFrame(resp *HogQLAPIResponse) (*data.Frame, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	frame := data.NewFrame("response")

	if len(resp.Columns) == 0 {
		return frame, nil
	}

	// Build typed fields based on ClickHouse type strings
	fields := make([]*data.Field, len(resp.Columns))
	for i, col := range resp.Columns {
		chType := ""
		if i < len(resp.Types) {
			chType = resp.Types[i]
		}
		fields[i] = newFieldForType(col, chType, len(resp.Results))
	}

	// Populate field values
	for rowIdx, row := range resp.Results {
		for colIdx, val := range row {
			if colIdx >= len(fields) {
				break
			}
			setFieldValue(fields[colIdx], rowIdx, val, resp.Types[colIdx])
		}
	}

	frame.Fields = fields
	return frame, nil
}

func newFieldForType(name string, chType string, length int) *data.Field {
	normalized := normalizeClickHouseType(chType)

	switch {
	case strings.Contains(normalized, "datetime") || strings.Contains(normalized, "date"):
		vals := make([]*time.Time, length)
		return data.NewField(name, nil, vals)
	case strings.Contains(normalized, "float") || strings.Contains(normalized, "decimal"):
		vals := make([]*float64, length)
		return data.NewField(name, nil, vals)
	case strings.Contains(normalized, "int") || strings.Contains(normalized, "uint"):
		vals := make([]*float64, length)
		return data.NewField(name, nil, vals)
	case strings.Contains(normalized, "bool"):
		vals := make([]*bool, length)
		return data.NewField(name, nil, vals)
	default:
		vals := make([]*string, length)
		return data.NewField(name, nil, vals)
	}
}

func setFieldValue(field *data.Field, idx int, val interface{}, chType string) {
	if val == nil {
		return // leave as nil pointer (null)
	}

	normalized := normalizeClickHouseType(chType)

	switch {
	case strings.Contains(normalized, "datetime") || strings.Contains(normalized, "date"):
		if s, ok := val.(string); ok {
			for _, layout := range []string{
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05.000Z",
				"2006-01-02",
			} {
				if t, err := time.Parse(layout, s); err == nil {
					field.Set(idx, &t)
					return
				}
			}
		}
		// Try numeric timestamp
		if f, ok := toFloat64(val); ok {
			t := time.Unix(int64(f), 0).UTC()
			field.Set(idx, &t)
		}
	case strings.Contains(normalized, "float") || strings.Contains(normalized, "decimal") ||
		strings.Contains(normalized, "int") || strings.Contains(normalized, "uint"):
		if f, ok := toFloat64(val); ok {
			field.Set(idx, &f)
		}
	case strings.Contains(normalized, "bool"):
		switch v := val.(type) {
		case bool:
			field.Set(idx, &v)
		case float64:
			b := v != 0
			field.Set(idx, &b)
		}
	default:
		s := fmt.Sprintf("%v", val)
		field.Set(idx, &s)
	}
}

func normalizeClickHouseType(chType string) string {
	t := strings.ToLower(chType)
	// Strip Nullable(...) wrapper
	if strings.HasPrefix(t, "nullable(") && strings.HasSuffix(t, ")") {
		t = t[len("nullable(") : len(t)-1]
	}
	return t
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		// PostHog sometimes returns numeric values as strings
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (d *PostHogDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	settings, secrets, err := LoadSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Unable to load settings",
		}, nil
	}

	if secrets.APIKey == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "API key is missing",
		}, nil
	}

	if settings.PostHogURL == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "PostHog URL is missing",
		}, nil
	}

	if settings.ProjectID == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Project ID is missing",
		}, nil
	}

	if err := d.client.TestConnection(ctx); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection test failed: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Successfully connected to PostHog",
	}, nil
}
