package downloader

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestDownloader_Download_Success(t *testing.T) {
	expectedBody := []byte("test response body")
	expectedContentType := "text/plain"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", expectedContentType)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(expectedBody)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	d := NewDownloader()
	u, _ := url.Parse(server.URL)

	body, contentType, err := d.Download(u)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(body) != string(expectedBody) {
		t.Errorf("expected body %s, got %s", expectedBody, body)
	}
	if contentType != expectedContentType {
		t.Errorf("expected content type %s, got %s", expectedContentType, contentType)
	}
}

func TestDownloader_Download_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	d := NewDownloader()
	u, _ := url.Parse(server.URL)

	_, _, err := d.Download(u)

	if err == nil {
		t.Fatal("expected error for 404 status")
	}
}

func TestDownloader_Download_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := NewDownloader()
	u, _ := url.Parse(server.URL)

	_, _, err := d.Download(u)

	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}

func TestDownloader_Download_InvalidURL(t *testing.T) {
	d := NewDownloader()
	u, _ := url.Parse("http://invalid-url-that-does-not-exist-12345.com")

	_, _, err := d.Download(u)

	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDownloader_Download_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewDownloader()
	u, _ := url.Parse(server.URL)

	_, _, err := d.Download(u)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestDownloader_Download_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewDownloader()
	u, _ := url.Parse(server.URL)

	body, _, err := d.Download(u)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty body, got %d bytes", len(body))
	}
}
