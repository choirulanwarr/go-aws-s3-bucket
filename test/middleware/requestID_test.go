package middleware_test

import (
	"github.com/gin-gonic/gin"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestID())

	// Add a test route that returns the API call ID
	router.GET("/test", func(c *gin.Context) {
		apiID, exists := c.Get(constant.RequestIDKey)
		if !exists {
			c.String(http.StatusInternalServerError, "no api_id")
			return
		}
		c.String(http.StatusOK, apiID.(string))
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	apiID := w.Body.String()
	if apiID == "" {
		t.Error("Expected API call ID, got empty response")
	}
	if apiID == "no api_id" {
		t.Error("API call ID was not set in context")
	}
	t.Logf("Generated API call ID: %s", apiID)
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestID())

	var ids []string
	router.GET("/test", func(c *gin.Context) {
		apiID, _ := c.Get(constant.RequestIDKey)
		ids = append(ids, apiID.(string))
		c.String(http.StatusOK, "ok")
	})

	// Make 3 requests
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i, w.Code)
		}
	}

	if len(ids) != 3 {
		t.Fatalf("Expected 3 IDs, got %d", len(ids))
	}

	// All IDs should be unique
	seen := make(map[string]bool)
	for i, id := range ids {
		if seen[id] {
			t.Errorf("Duplicate API call ID at request %d: %s", i, id)
		}
		seen[id] = true
	}
}

func TestRequestID_Format(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestID())

	router.GET("/test", func(c *gin.Context) {
		apiID, _ := c.Get(constant.RequestIDKey)
		c.String(http.StatusOK, apiID.(string))
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	apiID := w.Body.String()

	// Format should be: API_CALL_<timestamp>_<random>
	if len(apiID) < 20 {
		t.Errorf("API call ID too short: %s", apiID)
	}
}
