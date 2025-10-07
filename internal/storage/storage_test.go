package storage

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_Save(t *testing.T) {
	// Создаем временную директорию
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			t.Fatalf("failed to remove temp dir: %v", err)
		}
	}(tempDir)

	s := NewStorage(tempDir)

	tests := []struct {
		name        string
		urlStr      string
		content     []byte
		contentType string
		wantPath    string
		wantExt     string
	}{
		{
			name:        "save HTML file",
			urlStr:      "https://example.com/page",
			content:     []byte("<html>test</html>"),
			contentType: "text/html",
			wantPath:    filepath.Join(tempDir, "example.com", "page.html"),
			wantExt:     ".html",
		},
		{
			name:        "save CSS file",
			urlStr:      "https://example.com/styles/main.css",
			content:     []byte("body { margin: 0; }"),
			contentType: "text/css",
			wantPath:    filepath.Join(tempDir, "example.com", "styles", "main.css"),
			wantExt:     ".css",
		},
		{
			name:        "save JS file",
			urlStr:      "https://example.com/script.js",
			content:     []byte("console.log('test');"),
			contentType: "application/javascript",
			wantPath:    filepath.Join(tempDir, "example.com", "script.js"),
			wantExt:     ".js",
		},
		{
			name:        "save index page",
			urlStr:      "https://example.com/",
			content:     []byte("<html>index</html>"),
			contentType: "text/html",
			wantPath:    filepath.Join(tempDir, "example.com", "index.html"),
			wantExt:     ".html",
		},
		{
			name:        "save PNG image",
			urlStr:      "https://example.com/image.png",
			content:     []byte{0x89, 0x50, 0x4E, 0x47},
			contentType: "image/png",
			wantPath:    filepath.Join(tempDir, "example.com", "image.png"),
			wantExt:     ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			// Сохраняем файл
			err = s.Save(u, tt.content, tt.contentType)
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Проверяем существование файла
			if _, err := os.Stat(tt.wantPath); os.IsNotExist(err) {
				t.Errorf("Current files: %s", tempDir)
				t.Errorf("file does not exist at path: %s", tt.wantPath)
			}

			// Проверяем содержимое файла
			gotContent, err := os.ReadFile(tt.wantPath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			if string(gotContent) != string(tt.content) {
				t.Errorf("content mismatch:\ngot = %q\nwant = %q", gotContent, tt.content)
			}
		})
	}
}

func TestGetExtensionFromMIME(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"text/html", ".html"},
		{"text/html; charset=utf-8", ".html"},
		{"text/css", ".css"},
		{"application/javascript", ".js"},
		{"text/javascript", ".js"},
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"unknown/type", ".html"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := getExtensionFromMIME(tt.contentType)
			if got != tt.want {
				t.Errorf("getExtensionFromMIME(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
