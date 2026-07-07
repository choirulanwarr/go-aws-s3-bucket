package helper_test

import (
	"bytes"
	"go-aws-s3-bucket/app/helper"
	"mime/multipart"
	"testing"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1048576, "1.0 MB"},
		{"2.5 MB", 2621440, "2.5 MB"},
		{"1 GB", 1073741824, "1.0 GB"},
		{"1 TB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q; want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestDefaultMIME(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "application/octet-stream"},
		{"image jpeg", "image/jpeg", "image/jpeg"},
		{"image png", "image/png", "image/png"},
		{"application pdf", "application/pdf", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.DefaultMIME(tt.input)
			if result != tt.expected {
				t.Errorf("DefaultMIME(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueFilename(t *testing.T) {
	// Generate two filenames and verify they are different
	f1 := helper.GenerateUniqueFilename()
	f2 := helper.GenerateUniqueFilename()

	if f1 == "" {
		t.Error("GenerateUniqueFilename() returned empty string")
	}
	if f1 == f2 {
		t.Error("GenerateUniqueFilename() returned same value twice — should be unique")
	}
	if len(f1) < 15 {
		t.Errorf("GenerateUniqueFilename() = %q; expected at least 15 characters", f1)
	}
}

func TestIsAllowedFileType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		fileData    []byte
		expected    bool
	}{
		{
			name:        "valid JPEG",
			contentType: "image/jpeg",
			fileData:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, // JPEG magic bytes
			expected:    true,
		},
		{
			name:        "valid PNG",
			contentType: "image/png",
			fileData:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG magic bytes
			expected:    true,
		},
		{
			name:        "valid PDF",
			contentType: "application/pdf",
			fileData:    []byte{0x25, 0x50, 0x44, 0x46, 0x2D}, // %PDF-
			expected:    true,
		},
		{
			name:        "valid ZIP",
			contentType: "application/zip",
			fileData:    []byte{0x50, 0x4B, 0x03, 0x04}, // PK.. (ZIP magic)
			expected:    true,
		},
		{
			name:        "disallowed HTML",
			contentType: "text/html",
			fileData:    []byte("<html><body>test</body></html>"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a multipart file header from bytes
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			part, err := writer.CreateFormFile("file", "testfile")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}
			part.Write(tt.fileData)
			writer.Close()

			// Read back the multipart data as a file
			reader := multipart.NewReader(&buf, writer.Boundary())
			form, err := reader.ReadForm(10 << 20)
			if err != nil {
				t.Fatalf("Failed to read form: %v", err)
			}

			fileHeaders := form.File["file"]
			if len(fileHeaders) == 0 {
				t.Fatal("No file headers found")
			}

			file, err := fileHeaders[0].Open()
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			result := helper.IsAllowedFileType("TEST_API_CALL", file)
			if result != tt.expected {
				t.Errorf("IsAllowedFileType() = %v; want %v (content-type: %s)", result, tt.expected, tt.contentType)
			}
		})
	}
}
