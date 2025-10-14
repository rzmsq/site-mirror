package downloader

import (
	"errors"
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

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	body, contentType, err := d.Download(u, false)

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

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	_, _, err := d.Download(u, false)

	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected ErrTooManyAttempts, got %v", err)
	}
}

func TestDownloader_Download_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	_, _, err := d.Download(u, false)

	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected ErrTooManyAttempts, got %v", err)
	}
}

func TestDownloader_Download_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	_, _, err := d.Download(u, false)

	if err != nil {
		t.Fatalf("expected error")
	}
}

func TestDownloader_Download_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	body, _, err := d.Download(u, false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty body, got %d bytes", len(body))
	}
}

func TestDownloader_Download_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success after retry"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	d, _ := NewDownloader(u, "TestBot")

	body, _, err := d.Download(u, false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(body) != "success after retry" {
		t.Errorf("expected 'success after retry', got %s", string(body))
	}
}

func TestDownloader_Download_RobotsDisallowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("User-agent: TestBot\nDisallow: /blocked"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL + "/blocked")
	d, _ := NewDownloader(u, "TestBot")

	_, _, err := d.Download(u, true)

	if !errors.Is(err, ErrDisallowed) {
		t.Fatalf("expected ErrDisallowed, got %v", err)
	}
}
