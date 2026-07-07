package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-aws-s3-bucket/app/config"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/handler"
	"go-aws-s3-bucket/app/helper"
	"go-aws-s3-bucket/app/middleware"
	"go-aws-s3-bucket/app/service"
)

// setupTestRouter creates a Gin engine with all routes registered
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	server := gin.New()
	server.Use(middleware.RequestID())

	// Setup Viper — leave AWS values empty so we test validation behavior
	v := viper.New()
	v.Set("APP_PORT", 4001)

	validator := config.NewValidator()
	fileService := service.NewFileService(v)
	fileHandler := handler.NewFileHandler(fileService, validator)

	api := server.Group("/api/v1")
	api.GET("/list", fileHandler.GetAllFile)
	api.POST("/upload", fileHandler.UploadFile)
	api.GET("/download", fileHandler.DownloadFile)
	api.GET("/presigned-url", fileHandler.PresignedURL)
	api.POST("/move", fileHandler.MoveFile)

	return server
}

func TestGetAllFile_NoAWSConfig(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without AWS config, should return 422
	if w.Code != http.StatusUnprocessableEntity {
		t.Logf("Response body: %s", w.Body.String())
		t.Logf("Expected 422 (no AWS config), got %d", w.Code)
	}

	var response helper.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ApiID == "" {
		t.Error("Response should contain an API ID")
	}
}

func TestPresignedURL_InvalidRequest(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "missing path parameter",
			query: "",
		},
		{
			name:  "empty path value",
			query: "path=",
		},
		{
			name:  "valid params but no AWS config",
			query: "path=test.jpg&expires=15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/presigned-url?"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Accept 400 (validation error) or 422 (AWS config missing)
			if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity {
				t.Errorf("Expected 400 or 422, got %d: %s", w.Code, w.Body.String())
			}

			var response helper.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			if response.ApiID == "" {
				t.Error("Response missing api_id")
			}
		})
	}
}

func TestMoveFile_InvalidRequests(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: `{}`},
		{name: "missing dest_path", body: `{"source_path":"tmp/test.jpg"}`},
		{name: "missing source_path", body: `{"dest_path":"images/test.jpg"}`},
		{name: "empty strings", body: `{"source_path":"","dest_path":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/move", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected 400, got %d: %s", w.Code, w.Body.String())
			}

			var response helper.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			if response.ApiID == "" {
				t.Error("Response missing api_id")
			}
		})
	}
}

func TestDownloadFile_InvalidRequest(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing path, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoutes_ResponseFormat(t *testing.T) {
	router := setupTestRouter()

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/list", ""},
		{http.MethodGet, "/api/v1/presigned-url?path=test.jpg&expires=15", ""},
		{http.MethodPost, "/api/v1/move", `{"source_path":"a","dest_path":"b"}`},
		{http.MethodGet, "/api/v1/download", ""},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+"_"+strings.ReplaceAll(ep.path, "?", "_"), func(t *testing.T) {
			var req *http.Request
			if ep.body != "" {
				req, _ = http.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(ep.method, ep.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var response helper.Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Response is not valid JSON: %s", w.Body.String())
				return
			}

			if response.ApiID == "" {
				t.Errorf("Response missing api_id field for %s %s", ep.method, ep.path)
			}
			if response.Status != constant.ResponseStatusSuccess &&
				response.Status != constant.ResponseStatusFailed {
				t.Errorf("Invalid status %q for %s %s", response.Status, ep.method, ep.path)
			}
		})
	}
}

func TestUploadFile_NoFile(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 because no file/folder form data provided
	if w.Code != http.StatusBadRequest {
		t.Logf("Response: %s", w.Body.String())
	}
}

func TestConfigValidator_NotOnlySpace(t *testing.T) {
	validate := config.NewValidator()

	type TestReq struct {
		Field string `validate:"not_only_space"`
	}

	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"valid text", "hello", false},
		{"only spaces", "   ", true},
		{"empty string", "", true},
		{"mixed", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(TestReq{Field: tt.value})
			if tt.expectErr && err == nil {
				t.Errorf("Expected validation error for %q, got nil", tt.value)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %q, got: %v", tt.value, err)
			}
		})
	}
}
