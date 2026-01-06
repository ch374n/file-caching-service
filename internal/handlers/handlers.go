package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ch374n/file-downloader/internal/cache"
	"github.com/ch374n/file-downloader/internal/metrics"
	"github.com/ch374n/file-downloader/internal/storage"
)

// Response is the standard API response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// FileHandler handles file-related HTTP requests
type FileHandler struct {
	cache   cache.Cache
	storage storage.Storage
}

// NewFileHandler creates a new FileHandler with the given dependencies
func NewFileHandler(c cache.Cache, s storage.Storage) *FileHandler {
	return &FileHandler{
		cache:   c,
		storage: s,
	}
}

// Health handles health check requests
func (h *FileHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := map[string]string{
		"status": "healthy",
	}

	// Check cache (optional - doesn't affect overall health)
	if h.cache != nil {
		if err := h.cache.Ping(ctx); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	} else {
		health["redis"] = "disabled"
	}

	// Check storage (required - affects overall health)
	if err := h.storage.HealthCheck(ctx); err != nil {
		health["status"] = "unhealthy"
		health["r2"] = "unhealthy: " + err.Error()
		writeJSON(w, http.StatusServiceUnavailable, Response{
			Success: false,
			Message: "Service is unhealthy",
			Data:    health,
		})
		return
	}
	health["r2"] = "healthy"

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Service is healthy",
		Data:    health,
	})
}

// Root handles the root endpoint
func (h *FileHandler) Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "File Caching Service",
		Data: map[string]string{
			"version": "1.0.0",
		},
	})
}

// GetFile handles file retrieval requests
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("name")

	if filename == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "filename is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Check cache only if available
	if h.cache != nil {
		start := time.Now()
		data, found, err := h.cache.Get(ctx, filename)
		metrics.CacheOperationDuration.WithLabelValues("get").Observe(time.Since(start).Seconds())

		if err != nil {
			slog.Error("Cache error", "filename", filename, "error", err)
		}

		if found {
			metrics.CacheHitsTotal.Inc()
			slog.Info("Cache HIT", "filename", filename)
			writeFileResponse(w, filename, data)
			return
		}

		metrics.CacheMissesTotal.Inc()
		slog.Info("Cache MISS", "filename", filename)
	} else {
		slog.Info("Cache disabled, fetching from storage", "filename", filename)
	}

	// Fetch from storage
	start := time.Now()
	data, err := h.storage.GetObject(ctx, filename)
	duration := time.Since(start).Seconds()
	metrics.R2RequestDuration.WithLabelValues("get").Observe(duration)

	if err != nil {
		metrics.R2RequestsTotal.WithLabelValues("get", "error").Inc()
		slog.Error("Storage error", "filename", filename, "error", err)

		if ctx.Err() == context.DeadlineExceeded {
			writeJSON(w, http.StatusGatewayTimeout, Response{
				Success: false,
				Message: "Request timeout",
			})
			return
		}

		if isNotFoundError(err) {
			writeJSON(w, http.StatusNotFound, Response{
				Success: false,
				Message: "File not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to retrieve file",
		})
		return
	}

	metrics.R2RequestsTotal.WithLabelValues("get", "success").Inc()

	// Cache the file only if cache is available
	if h.cache != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			start := time.Now()
			if err := h.cache.Set(bgCtx, filename, data); err != nil {
				slog.Error("Failed to cache file", "filename", filename, "error", err)
			} else {
				slog.Info("Cached file", "filename", filename)
			}
			metrics.CacheOperationDuration.WithLabelValues("set").Observe(time.Since(start).Seconds())
		}()
	}

	writeFileResponse(w, filename, data)
}

// MetricsMiddleware wraps a handler to record HTTP metrics
func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapped, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		status := strconv.Itoa(wrapped.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		slog.Info("Request completed",
			"method", method,
			"path", path,
			"status", wrapped.statusCode,
			"duration_ms", duration*1000,
		)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func writeFileResponse(w http.ResponseWriter, filename string, data []byte) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "NoSuchKey") ||
		strings.Contains(err.Error(), "not found")
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding JSON response", "error", err)
	}
}
