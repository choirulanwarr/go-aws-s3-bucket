package helper_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"go-aws-s3-bucket/app/config"
	"go-aws-s3-bucket/app/helper"
)

func TestErrorValidationMessageGenerator(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		tag      string
		param    string
		expected string
	}{
		{
			name:     "required tag",
			field:    "folder",
			tag:      "required",
			param:    "",
			expected: "folder cannot be empty.",
		},
		{
			name:     "not_only_space tag",
			field:    "path",
			tag:      "not_only_space",
			param:    "",
			expected: "path value cannot be only 'space'.",
		},
		{
			name:     "unknown tag fallback",
			field:    "email",
			tag:      "email",
			param:    "",
			expected: "Field validation for 'email' failed on the 'email' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ErrorValidationMessageGenerator(tt.field, tt.tag, tt.param)
			if result != tt.expected {
				t.Errorf("ErrorValidationMessageGenerator(%q, %q, %q) = %q; want %q",
					tt.field, tt.tag, tt.param, result, tt.expected)
			}
		})
	}
}

func TestErrorValidationFormatter(t *testing.T) {
	validate := config.NewValidator()

	type TestRequest struct {
		Name  string `validate:"required"`
		Email string `validate:"required"`
	}

	req := TestRequest{}
	err := validate.Struct(req)

	if err == nil {
		t.Fatal("Expected validation errors, got nil")
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatal("Expected validator.ValidationErrors type")
	}

	formattedErrors := helper.ErrorValidationFormatter(validationErrors)

	if len(formattedErrors) == 0 {
		t.Error("Expected at least one formatted error, got 0")
	}

	// Verify structure of each formatted error
	for _, fe := range formattedErrors {
		if fe.Field == "" {
			t.Error("Formatted error field should not be empty")
		}
		if fe.Message == "" {
			t.Error("Formatted error message should not be empty")
		}
		if fe.Tag == "" {
			t.Error("Formatted error tag should not be empty")
		}
	}
}

func TestNotOnlySpaceValidator(t *testing.T) {
	validate := config.NewValidator()

	type TestStruct struct {
		Field string `validate:"not_only_space"`
	}

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"valid string", "hello", false},
		{"valid single char", "a", false},
		{"only spaces", "   ", true},
		{"empty string", "", true},
		{"spaces with text", "  hello  ", false},
		{"tab character", "\t", true},
		{"tab with text", "\thello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(TestStruct{Field: tt.input})
			if tt.expectErr && err == nil {
				t.Errorf("Expected validation error for %q, got nil", tt.input)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %q, got: %v", tt.input, err)
			}
		})
	}
}
