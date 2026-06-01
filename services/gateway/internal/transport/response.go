package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

type ResponseTransformer struct {
	stripHeaders     []string
	errorMapping     bool
	injectTraceID    bool
	maxResponseBytes int64
}

func NewResponseTransformer() *ResponseTransformer {
	return &ResponseTransformer{
		stripHeaders: []string{
			"Server",
			"X-Powered-By",
			"X-AspNet-Version",
		},
		errorMapping:     true,
		injectTraceID:    true,
		maxResponseBytes: 10 * 1024 * 1024,
	}
}

func (t *ResponseTransformer) TransformResponse(c *gin.Context, upstreamResp *http.Response) {
	t.stripSensitiveHeaders(upstreamResp)
	t.ensureContentType(upstreamResp)

	if t.injectTraceID {
		traceID := c.GetString(string("trace_id"))
		if traceID != "" {
			upstreamResp.Header.Set("X-Trace-ID", traceID)
		}
	}

	if upstreamResp.StatusCode >= 400 && t.errorMapping {
		t.mapUpstreamError(c, upstreamResp)
		return
	}

	t.writeResponse(c, upstreamResp)
}

func (t *ResponseTransformer) stripSensitiveHeaders(resp *http.Response) {
	for _, header := range t.stripHeaders {
		resp.Header.Del(header)
	}
}

func (t *ResponseTransformer) ensureContentType(resp *http.Response) {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		resp.Header.Set("Content-Type", "application/json")
	}
}

func (t *ResponseTransformer) mapUpstreamError(c *gin.Context, resp *http.Response) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, t.maxResponseBytes))
	if err != nil {
		observability.GetLogger().Error("failed to read upstream error body",
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"error_code": "UPSTREAM_ERROR",
			"message":    "upstream service error",
		})
		return
	}

	var upstreamErr map[string]interface{}
	if err := json.Unmarshal(body, &upstreamErr); err != nil {
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{
			"error_code": "UPSTREAM_ERROR",
			"message":    string(body),
		})
		return
	}

	mapped := gin.H{
		"error_code": "UPSTREAM_ERROR",
		"message":    "upstream service error",
		"status":     resp.StatusCode,
	}

	if code, ok := upstreamErr["error_code"]; ok {
		mapped["error_code"] = code
	}
	if msg, ok := upstreamErr["message"]; ok {
		mapped["message"] = msg
	}
	if details, ok := upstreamErr["details"]; ok {
		mapped["details"] = details
	}

	if traceID := c.GetString("trace_id"); traceID != "" {
		mapped["trace_id"] = traceID
	}

	c.AbortWithStatusJSON(resp.StatusCode, mapped)
}

func (t *ResponseTransformer) writeResponse(c *gin.Context, resp *http.Response) {
	for key, values := range resp.Header {
		for _, v := range values {
			c.Header(key, v)
		}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, t.maxResponseBytes))
	if err != nil {
		observability.GetLogger().Error("failed to read upstream response body",
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"error_code": "BAD_GATEWAY",
			"message":    "failed to read upstream response",
		})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func RequestTransformer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error_code": "INVALID_REQUEST_BODY",
					"message":    "cannot read request body",
				})
				return
			}

			if len(body) > 0 && isJSONContentType(c.GetHeader("Content-Type")) {
				body = normalizeJSONKeys(body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
		c.Next()
	}
}

func isJSONContentType(contentType string) bool {
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/*+json")
}

func normalizeJSONKeys(body []byte) []byte {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return body
	}

	normalized := make(map[string]interface{})
	for k, v := range raw {
		normalized[toSnakeCase(k)] = v
	}

	result, err := json.Marshal(normalized)
	if err != nil {
		return body
	}
	return result
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
