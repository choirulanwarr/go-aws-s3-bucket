package helper_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateApiCallID(t *testing.T) {
	callID1 := helper.GenerateApiCallID()
	callID2 := helper.GenerateApiCallID()

	if callID1 == "" {
		t.Error("GenerateApiCallID() returned empty string")
	}
	if callID1 == callID2 {
		t.Error("GenerateApiCallID() returned same value — should be unique per call")
	}
}

func TestResponseAPI(t *testing.T) {
	tests := []struct {
		name           string
		response       constant.ResponseMap
		additionalData []interface{}
		expectedCode   int
		expectedStatus string
		hasData        bool
	}{
		{
			name:           "success response without data",
			response:       constant.Res200Get,
			additionalData: nil,
			expectedCode:   http.StatusOK,
			expectedStatus: constant.ResponseStatusSuccess,
			hasData:        false,
		},
		{
			name:           "success response with data",
			response:       constant.Res200Get,
			additionalData: []interface{}{map[string]string{"key": "value"}},
			expectedCode:   http.StatusOK,
			expectedStatus: constant.ResponseStatusSuccess,
			hasData:        true,
		},
		{
			name:           "bad request response",
			response:       constant.Res400InvalidPayload,
			additionalData: nil,
			expectedCode:   http.StatusBadRequest,
			expectedStatus: constant.ResponseStatusFailed,
			hasData:        false,
		},
		{
			name:           "unprocessable entity response",
			response:       constant.Res422SomethingWentWrong,
			additionalData: nil,
			expectedCode:   http.StatusUnprocessableEntity,
			expectedStatus: constant.ResponseStatusFailed,
			hasData:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set API call ID in context (simulating RequestID middleware)
			c.Set(constant.RequestIDKey, "API_CALL_TEST_1234567")

			helper.ResponseAPI(c, tt.response, tt.additionalData...)

			if w.Code != tt.expectedCode {
				t.Errorf("ResponseAPI() status = %d; want %d", w.Code, tt.expectedCode)
			}

			var response helper.Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.ApiID != "API_CALL_TEST_1234567" {
				t.Errorf("Response API ID = %q; want %q", response.ApiID, "API_CALL_TEST_1234567")
			}
			if response.Status != tt.expectedStatus {
				t.Errorf("Response status = %q; want %q", response.Status, tt.expectedStatus)
			}
			if tt.hasData && response.Data == nil {
				t.Error("Expected data in response but got nil")
			}
		})
	}
}

func TestResponseAPIWithoutRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Don't set API call ID — should auto-generate one
	helper.ResponseAPI(c, constant.Res200Get)

	if w.Code != http.StatusOK {
		t.Errorf("ResponseAPI() status = %d; want %d", w.Code, http.StatusOK)
	}

	var response helper.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ApiID == "" {
		t.Error("Expected auto-generated API ID but got empty string")
	}
}
