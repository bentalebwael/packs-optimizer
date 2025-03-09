package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter captures the status code and body of the response
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// Write captures the response body
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString captures the response string
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger returns a middleware that logs request and response details
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Create a buffer to store the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the body for subsequent middleware/handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create custom response writer
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK, // Default status code
		}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Prepare fields for structured logging
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status_code", w.statusCode),
			zap.Duration("duration", duration),
		}

		// Add request body if present and not too large
		if len(requestBody) > 0 && len(requestBody) < 1024*10 { // Limit to 10KB
			fields = append(fields, zap.String("request_body", string(requestBody)))
		}

		// Add response body if present and not too large
		responseBody := w.body.String()
		if len(responseBody) > 0 && len(responseBody) < 1024*10 { // Limit to 10KB
			fields = append(fields, zap.String("response_body", responseBody))
		}

		// Log at appropriate level based on status code
		if w.statusCode >= 500 {
			log.Error("Request completed with server error", fields...)
		} else if w.statusCode >= 400 {
			log.Warn("Request completed with client error", fields...)
		} else {
			log.Info("Request completed successfully", fields...)
		}
	}
}
