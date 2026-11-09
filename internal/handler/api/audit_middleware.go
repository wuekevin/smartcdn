package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/infra/logger"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

// auditableMethod returns true for methods that mutate state.
func auditableMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

type responseCapture struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.status = code
	rc.ResponseWriter.WriteHeader(code)
}

// AuditLoggerMiddleware records mutating API calls to audit_logs asynchronously (AD-15).
func AuditLoggerMiddleware(auditRepo *repo.AuditRepo) gin.HandlerFunc {
	log := logger.Named("audit")
	return func(c *gin.Context) {
		if !auditableMethod(c.Request.Method) {
			c.Next()
			return
		}

		// 读取请求体（供审计记录），然后放回供后续 handler 使用
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 捕获响应体
		capture := &responseCapture{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}
		c.Writer = capture

		c.Next()

		// 只记录成功的变更操作（2xx）
		if capture.status < 200 || capture.status >= 300 {
			return
		}

		userID, _ := GetUserID(c)
		action := buildAction(c.Request.Method, c.FullPath())

		var afterStateStr *string
		if capture.body.Len() > 0 && capture.body.Len() <= 4096 {
			s := compactJSON(capture.body.Bytes())
			afterStateStr = &s
		}

		entry := &entity.AuditLog{
			OperatorIP:   c.ClientIP(),
			Action:       action,
			ResourceType: extractResourceType(c.FullPath()),
			AfterState:   afterStateStr,
		}
		if userID != 0 {
			entry.OperatorID = &userID
		}

		go func() {
			if err := auditRepo.Create(context.Background(), entry); err != nil {
				log.Sugar().Warnw("audit log write failed", "error", err)
			}
		}()
	}
}

func buildAction(method, path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		// /api/v1/dns-records → dns_records.create
		resource := strings.ReplaceAll(parts[2], "-", "_")
		switch method {
		case http.MethodPost:
			return resource + ".create"
		case http.MethodPut, http.MethodPatch:
			return resource + ".update"
		case http.MethodDelete:
			return resource + ".delete"
		}
	}
	return strings.ToLower(method) + ":" + path
}

func extractResourceType(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return strings.ReplaceAll(parts[2], "-", "_")
	}
	return ""
}

func compactJSON(b []byte) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return string(b)
	}
	return buf.String()
}
